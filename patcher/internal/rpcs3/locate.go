package rpcs3

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"anamod/internal/paths"
	"go.yaml.in/yaml/v3"
)

type Location struct {
	Folder    string
	Root      string
	ConfigDir string
	DevHDD0   string
}

func (l Location) PatchConfigPath() string {
	return filepath.Join(l.ConfigDir, "patch_config.yml")
}

func (l Location) PatchesDir() string {
	return filepath.Join(l.Root, "patches")
}

var (
	ErrNoExe    = errors.New("this folder has no rpcs3 executable. Choose the folder RPCS3 runs from")
	ErrNoConfig = errors.New("this folder has no RPCS3 settings (config.yml). Start RPCS3 " +
		"once, or choose its config folder")
)

func FromFolder(folder, goos string) (Location, error) {
	folder = filepath.Clean(folder)
	loc := Location{Folder: folder}
	if goos == "windows" {
		if !paths.IsFile(filepath.Join(folder, "rpcs3.exe")) {
			return Location{}, ErrNoExe
		}
		loc.Root = folder
		if paths.IsDir(filepath.Join(folder, "portable")) {
			loc.Root = filepath.Join(folder, "portable")
		}
		loc.ConfigDir = filepath.Join(loc.Root, "config")
	} else {
		if !paths.IsFile(filepath.Join(folder, "config.yml")) {
			return Location{}, ErrNoConfig
		}
		loc.Root = folder
		loc.ConfigDir = folder
	}
	loc.DevHDD0 = devHDD0(loc)
	return loc, nil
}

func devHDD0(loc Location) string {
	fallback := filepath.Join(loc.Root, "dev_hdd0")
	data, err := os.ReadFile(filepath.Join(loc.ConfigDir, "vfs.yml"))
	if err != nil {
		return fallback
	}
	var vfs map[string]any
	if err := yaml.Unmarshal(data, &vfs); err != nil {
		return fallback
	}
	value, ok := vfs["/dev_hdd0/"].(string)
	if !ok || value == "" {
		return fallback
	}
	value = strings.ReplaceAll(value, "$(EmulatorDir)", loc.Root+string(filepath.Separator))
	return filepath.Clean(value)
}

func Candidates(goos, home, xdgConfig string) []string {
	var found []string
	if goos == "windows" {
		for _, folder := range []string{
			filepath.Join(home, "Desktop", "rpcs3"),
			filepath.Join(home, "Downloads", "rpcs3"),
			filepath.Join(home, "rpcs3"),
			filepath.Join(home, "AppData", "Roaming", "emudeck", "Emulators", "RPCS3"),
			`C:\rpcs3`,
			`C:\Emulators\rpcs3`,
		} {
			if paths.IsFile(filepath.Join(folder, "rpcs3.exe")) {
				found = append(found, folder)
			}
		}
		return found
	}
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}
	for _, folder := range []string{
		filepath.Join(xdgConfig, "rpcs3"),
		filepath.Join(home, ".var", "app", "net.rpcs3.RPCS3", "config", "rpcs3"),
	} {
		if paths.IsFile(filepath.Join(folder, "config.yml")) {
			found = append(found, folder)
		}
	}
	return found
}
