package install

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"anamod/internal/atomicfile"
	"anamod/internal/bps"
	"anamod/internal/datapack"
	"anamod/internal/paths"
	"anamod/internal/xenfile"
)

type installPlan struct {
	root       string
	pack       *datapack.Pack
	game       datapack.Game
	old        *Record
	log        io.Writer
	record     *Record
	newBackups []fileWrite
	outputs    []fileWrite
	removals   []string
	tocRows    []xenfile.TocRow
	tablePath  string
	table      []byte
}

func Install(root string, pack *datapack.Pack, game datapack.Game, options []string,
	log io.Writer) (*Record, error) {
	result, err := checkInstallable(root, pack, game)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(log, "Installing AnaMod %s into %s (found: %s)\n",
		pack.Manifest.AnamodVersion, root, result.State)
	plan := newInstallPlan(root, pack, game, options, result.Record, log)
	steps := []func() error{plan.addChanged, plan.addAdded, plan.addTable,
		plan.restoreOld}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}
	return plan.commit()
}

func checkInstallable(root string, pack *datapack.Pack, game datapack.Game) (Result, error) {
	result, err := Check(root, pack.Manifest)
	if err != nil {
		return result, err
	}
	if result.Game.ID != game.ID {
		return result, fmt.Errorf("this folder holds %s, not %s", result.Game.Name, game.Name)
	}
	if result.State != Vanilla && result.State != Installed {
		return result, errors.New("AnaMod can only be installed on clean game files, and some " +
			"files aren't")
	}
	return result, nil
}

func newInstallPlan(root string, pack *datapack.Pack, game datapack.Game, options []string,
	old *Record, log io.Writer) *installPlan {
	record := &Record{
		AnamodVersion: pack.Manifest.AnamodVersion,
		Options:       options,
		Written:       map[string]string{},
		Backups:       map[string]string{},
	}
	if old != nil {
		record.RPCS3Dir = old.RPCS3Dir
		record.OfficialFPSWasOn = old.OfficialFPSWasOn
		maps.Copy(record.Backups, old.Backups)
	}
	return &installPlan{root: root, pack: pack, game: game, old: old, log: log, record: record}
}

func (p *installPlan) readVanillaAndKeep(path, vanillaHash string) ([]byte, error) {
	vanilla, haveBackup, err := readVanilla(p.root, path, vanillaHash)
	if err != nil {
		return nil, err
	}
	if !haveBackup {
		p.newBackups = append(p.newBackups, fileWrite{backupPath(p.root, path),
			BackupDir + "/files/" + path, vanilla})
	}
	p.record.Backups[path] = vanillaHash
	return vanilla, nil
}

func (p *installPlan) addChanged() error {
	for _, file := range p.game.Changed {
		if err := p.addChangedFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (p *installPlan) addChangedFile(file datapack.ChangedFile) error {
	vanilla, err := p.readVanillaAndKeep(file.Path, file.VanillaSHA256)
	if err != nil {
		return err
	}
	if file.Table {
		p.tablePath, p.table = file.Path, vanilla
		return nil
	}
	diff, err := p.pack.ReadFile(file.Diff)
	if err != nil {
		return err
	}
	patched, err := applyDiff(vanilla, diff, file)
	if err != nil {
		return err
	}
	fmt.Fprintf(p.log, "Patched %s: %s\n", file.Path, file.ResultSHA256)
	written := patched
	if file.Compressed {
		if written, err = compress(patched, file.Path); err != nil {
			return err
		}
		p.tocRows = append(p.tocRows, xenfile.TocRow{Key: file.TocKey,
			Inflated: uint32(len(patched)), Stored: uint32(len(written))})
	}
	p.outputs = append(p.outputs, fileWrite{gamePath(p.root, file.Path), file.Path, written})
	p.record.Written[file.Path] = bytesSHA256(written)
	return nil
}

func (p *installPlan) addAdded() error {
	for _, file := range p.game.Added {
		if err := p.addAddedFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (p *installPlan) addAddedFile(file datapack.AddedFile) error {
	data, err := p.pack.ReadFile(file.File)
	if err != nil {
		return err
	}
	if found := bytesSHA256(data); found != file.SHA256 {
		return fmt.Errorf("the data pack's copy of %s is damaged: %s, expected %s",
			file.Path, found, file.SHA256)
	}
	if file.Compressed {
		row, err := addedRow(data, file)
		if err != nil {
			return err
		}
		p.tocRows = append(p.tocRows, row)
	}
	p.outputs = append(p.outputs, fileWrite{gamePath(p.root, file.Path), file.Path, data})
	p.record.Written[file.Path] = file.SHA256
	p.record.Added = append(p.record.Added, file.Path)
	return nil
}

func (p *installPlan) addTable() error {
	if p.tablePath == "" {
		return nil
	}
	updated, err := xenfile.UpdateTOC(p.table, p.tocRows)
	if err != nil {
		return fmt.Errorf("could not update %s: %w", p.tablePath, err)
	}
	fmt.Fprintf(p.log, "Updated %s: %d rows\n", p.tablePath, len(p.tocRows))
	p.outputs = append(p.outputs, fileWrite{gamePath(p.root, p.tablePath), p.tablePath, updated})
	p.record.Written[p.tablePath] = bytesSHA256(updated)
	return nil
}

func (p *installPlan) restoreOld() error {
	if p.old == nil {
		return nil
	}
	for _, path := range slices.Sorted(maps.Keys(p.old.Written)) {
		if _, ok := p.record.Written[path]; ok {
			continue
		}
		if slices.Contains(p.old.Added, path) {
			p.removals = append(p.removals, path)
			continue
		}
		vanilla, err := readBackup(p.root, path, p.old.Backups[path])
		if err != nil {
			return err
		}
		p.outputs = append(p.outputs, fileWrite{gamePath(p.root, path), path, vanilla})
	}
	return nil
}

func (p *installPlan) commit() (*Record, error) {
	recordData, err := json.MarshalIndent(p.record, "", "  ")
	if err != nil {
		return nil, err
	}

	createdBackupDir := !paths.IsDir(filepath.Join(p.root, BackupDir))
	changes := &changeLog{log: p.log}
	err = changes.apply(p.root, append(p.newBackups, p.outputs...), p.removals)
	if err == nil {
		err = changes.write(fileWrite{recordPath(p.root), BackupDir + "/" + RecordName,
			recordData})
	}
	if err != nil {
		fmt.Fprintf(p.log, "Install failed: %v. Restoring the original files.\n", err)
		changes.rollback()
		if createdBackupDir {
			os.RemoveAll(filepath.Join(p.root, BackupDir))
		}
		return nil, err
	}
	fmt.Fprintf(p.log, "Installed AnaMod %s\n", p.record.AnamodVersion)
	return p.record, nil
}

func applyDiff(vanilla, diff []byte, file datapack.ChangedFile) ([]byte, error) {
	source := vanilla
	if file.Compressed {
		inflated, err := xenfile.Inflate(vanilla)
		if err != nil {
			return nil, fmt.Errorf("could not patch %s: %w", file.Path, err)
		}
		source = inflated
	}
	patched, err := bps.Apply(source, diff)
	if err != nil {
		return nil, fmt.Errorf("could not patch %s: %w", file.Path, err)
	}
	if found := bytesSHA256(patched); found != file.ResultSHA256 {
		return nil, fmt.Errorf("patching %s gave %s, expected %s", file.Path, found,
			file.ResultSHA256)
	}
	return patched, nil
}

func compress(contents []byte, path string) ([]byte, error) {
	packed, err := xenfile.Deflate(contents)
	if err != nil {
		return nil, fmt.Errorf("could not compress %s: %w", path, err)
	}
	if err := xenfile.CheckInPlace(packed); err != nil {
		return nil, fmt.Errorf("%s would not load: %w", path, err)
	}
	return packed, nil
}

func addedRow(packed []byte, file datapack.AddedFile) (xenfile.TocRow, error) {
	inflated, err := xenfile.Inflate(packed)
	if err != nil {
		return xenfile.TocRow{}, fmt.Errorf("the data pack's copy of %s is damaged: %w",
			file.Path, err)
	}
	if err := xenfile.CheckInPlace(packed); err != nil {
		return xenfile.TocRow{}, fmt.Errorf("%s would not load: %w", file.Path, err)
	}
	return xenfile.TocRow{Key: file.TocKey, Inflated: uint32(len(inflated)),
		Stored: uint32(len(packed))}, nil
}

func Uninstall(root string, game datapack.Game, log io.Writer) error {
	record, err := ReadRecord(root)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("AnaMod isn't installed in this folder")
	}
	fmt.Fprintf(log, "Uninstalling AnaMod %s from %s\n", record.AnamodVersion, root)
	vanillaHashes := map[string]string{}
	for _, file := range game.Changed {
		vanillaHashes[file.Path] = file.VanillaSHA256
	}
	maps.Copy(vanillaHashes, record.Backups)
	var restores []fileWrite
	var removals []string
	for _, path := range slices.Sorted(maps.Keys(record.Written)) {
		if slices.Contains(record.Added, path) {
			removals = append(removals, path)
			continue
		}
		vanilla, err := readBackup(root, path, vanillaHashes[path])
		if err != nil {
			return err
		}
		restores = append(restores, fileWrite{gamePath(root, path), path, vanilla})
	}
	changes := &changeLog{log: log}
	if err := changes.apply(root, restores, removals); err != nil {
		fmt.Fprintf(log, "Uninstall failed... %v. Restoring the original files.\n", err)
		changes.rollback()
		return err
	}
	if err := os.RemoveAll(filepath.Join(root, BackupDir)); err != nil {
		return fmt.Errorf("the game files are back to vanilla, but %s could not be "+
			"removed: %w", BackupDir, err)
	}
	fmt.Fprintf(log, "Removed %s. The game files are vanilla again.\n", BackupDir)
	return nil
}

func readVanilla(root, relative, vanillaHash string) (data []byte, haveBackup bool,
	err error) {
	if _, err := os.Stat(backupPath(root, relative)); err == nil {
		data, err := readBackup(root, relative, vanillaHash)
		return data, true, err
	}
	data, err = os.ReadFile(gamePath(root, relative))
	if err != nil {
		return nil, false, fmt.Errorf("could not read %s: %w", relative, err)
	}
	if found := bytesSHA256(data); found != vanillaHash {
		return nil, false, fmt.Errorf("%s is not the original file: %s, expected %s",
			relative, found, vanillaHash)
	}
	return data, false, nil
}

func readBackup(root, relative, vanillaHash string) ([]byte, error) {
	data, err := os.ReadFile(backupPath(root, relative))
	if err != nil {
		return nil, fmt.Errorf("the backup of %s could not be read: %w", relative, err)
	}
	if found := bytesSHA256(data); vanillaHash != "" && found != vanillaHash {
		return nil, fmt.Errorf("the backup of %s is not the original file: %s, expected %s",
			relative, found, vanillaHash)
	}
	return data, nil
}

type fileWrite struct {
	path string
	name string
	data []byte
}

type changeLog struct {
	log  io.Writer
	undo []undoStep
}

type undoStep struct {
	path    string
	existed bool
	prior   []byte
}

func (c *changeLog) remember(path string) error {
	prior, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		c.undo = append(c.undo, undoStep{path: path})
		return nil
	}
	if err != nil {
		return err
	}
	c.undo = append(c.undo, undoStep{path: path, existed: true, prior: prior})
	return nil
}

func (c *changeLog) apply(root string, writes []fileWrite, removals []string) error {
	for _, write := range writes {
		if err := c.write(write); err != nil {
			return err
		}
	}
	for _, path := range removals {
		if err := c.remove(gamePath(root, path), path); err != nil {
			return err
		}
	}
	return nil
}

func (c *changeLog) write(w fileWrite) error {
	if err := c.remember(w.path); err != nil {
		return fmt.Errorf("could not read %s: %w", w.name, err)
	}
	if err := atomicfile.Write(w.path, w.data); err != nil {
		return fmt.Errorf("could not write %s: %w", w.name, err)
	}
	fmt.Fprintf(c.log, "Wrote %s (%s)\n", w.name, bytesSHA256(w.data))
	return nil
}

func (c *changeLog) remove(path, name string) error {
	if err := c.remember(path); err != nil {
		return fmt.Errorf("could not read %s: %w", name, err)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("could not delete %s: %w", name, err)
	}
	fmt.Fprintf(c.log, "Deleted %s\n", name)
	return nil
}

func (c *changeLog) rollback() {
	for _, step := range slices.Backward(c.undo) {
		var err error
		if step.existed {
			err = atomicfile.Write(step.path, step.prior)
		} else {
			err = os.Remove(step.path)
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
		}
		if err != nil {
			fmt.Fprintf(c.log, "Could not put back %s: %v\n", step.path, err)
		} else {
			fmt.Fprintf(c.log, "Put back %s\n", step.path)
		}
	}
}
