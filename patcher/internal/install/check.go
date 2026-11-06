package install

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"anamod/internal/datapack"
	"anamod/internal/paths"
	"anamod/internal/sfo"
	"anamod/internal/xex"
)

const xexName = "default.xex"

var ErrNotAGameFolder = errors.New("this isn't a game folder. Choose the folder that " +
	"holds PS3_GAME (PS3) or default.xex (360). ISOs and GoD are not supported.")

type State int

const (
	NoGame State = iota
	Vanilla
	Installed
	ChangedSinceInstall
	SomethingElse
)

func (s State) String() string {
	switch s {
	case NoGame:
		return "No Game"
	case Vanilla:
		return "Vanilla"
	case Installed:
		return "Installed"
	case ChangedSinceInstall:
		return "Changed since install"
	case SomethingElse:
		return "Something else"
	}
	return fmt.Sprintf("State(%d)", int(s))
}

type Mismatch struct {
	Path     string
	Expected string
	Found    string
	Missing  bool
}

type Result struct {
	Game       datapack.Game
	State      State
	Mismatches []Mismatch
	Record     *Record
}

type WrongRegionError struct {
	Found    string
	Accepted []string
}

func (e *WrongRegionError) Error() string {
	return "wrong region: " + e.Detail()
}

func (e *WrongRegionError) Detail() string {
	return fmt.Sprintf("this is %s, AnaMod needs %s", e.Found, strings.Join(e.Accepted, " or "))
}

func FindRoot(dropped string) (string, error) {
	path := filepath.Clean(dropped)
	for range 3 {
		if paths.IsDir(filepath.Join(path, "PS3_GAME")) || paths.IsFile(filepath.Join(path, xexName)) {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "", ErrNotAGameFolder
}

func readTitleID(root string) (string, error) {
	if !paths.IsDir(filepath.Join(root, "PS3_GAME")) && paths.IsFile(filepath.Join(root, xexName)) {
		data, err := os.ReadFile(filepath.Join(root, xexName))
		if err != nil {
			return "", fmt.Errorf("could not read %s: %w", xexName, err)
		}
		return xex.TitleID(data)
	}
	data, err := os.ReadFile(filepath.Join(root, "PS3_GAME", "PARAM.SFO"))
	if err != nil {
		return "", fmt.Errorf("could not read PS3_GAME/PARAM.SFO: %w", err)
	}
	return sfo.TitleID(data)
}

func Check(root string, manifest datapack.Manifest) (Result, error) {
	titleID, err := readTitleID(root)
	if err != nil {
		return Result{}, err
	}
	var accepted []string
	gameIndex := -1
	for i, game := range manifest.Games {
		accepted = append(accepted, game.TitleID)
		if game.TitleID == titleID && gameIndex < 0 {
			gameIndex = i
		}
	}
	if gameIndex < 0 {
		return Result{}, &WrongRegionError{Found: titleID, Accepted: accepted}
	}
	result := Result{Game: manifest.Games[gameIndex]}
	record, err := ReadRecord(root)
	if err != nil {
		return Result{}, err
	}
	result.Record = record
	eboot, err := checkFile(root, result.Game.Eboot.Path, result.Game.Eboot.SHA256)
	if err != nil {
		return Result{}, err
	}
	if eboot != nil {
		result.Mismatches = append(result.Mismatches, *eboot)
	}
	if record == nil {
		for _, file := range result.Game.Changed {
			mismatch, err := checkFile(root, file.Path, []string{file.VanillaSHA256})
			if err != nil {
				return Result{}, err
			}
			if mismatch != nil {
				result.Mismatches = append(result.Mismatches, *mismatch)
			}
		}
		result.State = Vanilla
		if len(result.Mismatches) > 0 {
			result.State = SomethingElse
		}
		return result, nil
	}
	for _, path := range slices.Sorted(maps.Keys(record.Written)) {
		mismatch, err := checkFile(root, path, []string{record.Written[path]})
		if err != nil {
			return Result{}, err
		}
		if mismatch != nil {
			result.Mismatches = append(result.Mismatches, *mismatch)
		}
	}
	result.State = Installed
	if len(result.Mismatches) > 0 {
		result.State = ChangedSinceInstall
	}
	return result, nil
}

func checkFile(root, relative string, expected []string) (*Mismatch, error) {
	found, err := fileSHA256(gamePath(root, relative))
	missing := errors.Is(err, fs.ErrNotExist)
	if err != nil && !missing {
		return nil, fmt.Errorf("could not read %s: %w", relative, err)
	}
	if !missing && slices.Contains(expected, found) {
		return nil, nil
	}
	return &Mismatch{Path: relative, Expected: strings.Join(expected, " or "), Found: found,
		Missing: missing}, nil
}
