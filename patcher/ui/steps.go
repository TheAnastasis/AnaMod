package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"anamod/internal/cache"
	"anamod/internal/datapack"
	"anamod/internal/install"
	"anamod/internal/rpcs3"
)

const usrdirPrefix = "PS3_GAME/USRDIR/"

func pickPack(embedded, downloaded []byte, patcherVersion string) (*datapack.Pack, error) {
	pack, err := datapack.Open(embedded)
	if err != nil {
		return nil, err
	}
	if newer, err := openUpdate(downloaded, patcherVersion, pack); err == nil {
		return newer, nil
	}
	return pack, nil
}

func openUpdate(downloaded []byte, patcherVersion string,
	current *datapack.Pack) (*datapack.Pack, error) {
	pack, err := datapack.Open(downloaded)
	if err != nil {
		return nil, err
	}
	if datapack.CompareVersions(pack.Manifest.MinPatcher, patcherVersion) > 0 {
		return nil, datapack.ErrNeedsNewerPatcher
	}
	if datapack.CompareVersions(pack.Manifest.AnamodVersion,
		current.Manifest.AnamodVersion) <= 0 {
		return nil, fmt.Errorf("the downloaded data pack holds AnaMod %s, which is not newer "+
			"than %s", pack.Manifest.AnamodVersion, current.Manifest.AnamodVersion)
	}
	return pack, nil
}

func doInstall(root string, res install.Result, pack *datapack.Pack, loc *rpcs3.Location,
	chosen []string, log io.Writer) (string, error) {
	game := res.Game
	previous := res.Record
	record, err := install.Install(root, pack, game, chosen, log)
	if err != nil {
		return "", err
	}
	header := fmt.Sprintf("AnaMod %s is installed!", record.AnamodVersion)
	lines := []string{install.BackupDir + " has the original files, in case you want to " +
		"uninstall the mod."}
	if game.Platform == "xen" {
		lines = append(lines, game.InstallNotes...)
		return installSummary(header, lines), nil
	}
	if loc == nil {
		fmt.Fprintln(log, "No RPCS3 folder chosen: RPCS3's patches were not set up and "+
			"its hard drive cache was not refreshed.")
		lines = append(lines, "RPCS3 wasn't set up, so the RPCS3 patches are off. Choose your "+
			"RPCS3 folder and press Update to set them up.")
		return installSummary(header, lines), nil
	}
	if previous != nil && previous.RPCS3Dir != "" && previous.RPCS3Dir != loc.Folder {
		removeFromEarlierRPCS3(previous, game, log)
	}
	patchFile, err := pack.ReadFile(game.RPCS3.PatchFile)
	if err != nil {
		return "", rpcs3NotSetUp(err)
	}
	wasOn, warnings, err := rpcs3.Apply(*loc, game.RPCS3, patchFile, chosen, log)
	if err != nil {
		return "", rpcs3NotSetUp(err)
	}
	if previous == nil || previous.RPCS3Dir != loc.Folder {
		record.OfficialFPSWasOn = wasOn
	}
	record.RPCS3Dir = loc.Folder
	if err := install.WriteRecord(root, record); err != nil {
		return "", rpcs3NotSetUp(err)
	}
	lines = append(lines, "RPCS3 Patches Installed: "+optionLabels(game.RPCS3, chosen))
	cacheWarnings := refreshCache(root, game, *loc, log)
	lines = append(lines, cacheWarnings...)
	lines = append(lines, warnings...)
	return installSummary(header, lines), nil
}

func installSummary(header string, bullets []string) string {
	var summary strings.Builder
	summary.WriteString(header)
	for _, bullet := range bullets {
		summary.WriteString("\n- " + bullet)
	}
	return summary.String()
}

func rpcs3NotSetUp(cause error) error {
	return &playerError{title: "RPCS3 Not Set Up", message: "AnaMod's game files are " +
		"installed, but RPCS3 couldn't be set up. Close RPCS3 and press Update to try again.",
		cause: cause}
}

func doUninstall(root string, res install.Result, loc *rpcs3.Location, log io.Writer) error {
	record := res.Record
	if record == nil {
		return errors.New("AnaMod isn't installed in this folder")
	}
	if err := install.Uninstall(root, res.Game, log); err != nil {
		return err
	}
	if loc == nil {
		fmt.Fprintln(log, "No RPCS3 folder: RPCS3's patches and cache were left as they are.")
		return nil
	}
	if err := rpcs3.Remove(*loc, res.Game.RPCS3, record.OfficialFPSWasOn, log); err != nil {
		return &playerError{title: "RPCS3 Not Updated", message: "Your script files are " +
			"the original versions again, but AnaMod's patches couldn't be unticked in RPCS3. Untick them " +
			"in RPCS3's Patch Manager.", cause: err}
	}
	warnings := refreshCache(root, res.Game, *loc, log)
	if len(warnings) > 0 {
		return &playerError{title: "RPCS3 Wasn't Updated", message: "Your script files are " +
			"the original versions again, but some of RPCS3's cached copies couldn't be updated, so the " +
			"game may still load AnaMod's versions of them.",
			cause: errors.New(strings.Join(warnings, " "))}
	}
	return nil
}

func refreshCache(root string, game datapack.Game, loc rpcs3.Location,
	log io.Writer) []string {
	var warnings []string
	for _, file := range game.Changed {
		if file.CacheRecord == "" || !strings.HasPrefix(file.Path, usrdirPrefix) {
			continue
		}
		relative := strings.TrimPrefix(file.Path, usrdirPrefix)
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file.Path)))
		cached := false
		if err == nil {
			cached, err = cache.Refresh(loc.DevHDD0, game.TitleID, relative, file.CacheRecord,
				content)
		}
		switch {
		case err != nil:
			fmt.Fprintf(log, "Cache: %v\n", err)
			warnings = append(warnings, fmt.Sprintf("RPCS3's cached copy of %s couldn't be "+
				"updated (%v), so the game may still load the old copy.", file.Path, err))
		case cached:
			fmt.Fprintf(log, "Refreshed RPCS3's cached copy of %s\n", file.Path)
		default:
			fmt.Fprintf(log, "RPCS3 keeps no cached copy of %s yet\n", file.Path)
		}
	}
	return warnings
}

func removeFromEarlierRPCS3(previous *install.Record, game datapack.Game, log io.Writer) {
	old, err := rpcs3.FromFolder(previous.RPCS3Dir, runtime.GOOS)
	if err == nil {
		err = rpcs3.Remove(old, game.RPCS3, previous.OfficialFPSWasOn, log)
	}
	if err != nil {
		fmt.Fprintf(log, "Could not untick AnaMod in the earlier RPCS3 %s: %v\n",
			previous.RPCS3Dir, err)
	}
}

func optionLabels(info datapack.RPCS3Info, chosen []string) string {
	var labels []string
	for _, option := range info.Options {
		if slices.Contains(chosen, option.ID) {
			labels = append(labels, option.Label)
		}
	}
	if len(labels) == 0 {
		return "no options ticked"
	}
	return strings.Join(labels, ", ")
}
