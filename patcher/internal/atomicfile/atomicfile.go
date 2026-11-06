package atomicfile

import (
	"os"
	"path/filepath"
	"runtime"
)

func Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".anamod-tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		os.Remove(temporary)
		return err
	}
	if runtime.GOOS == "windows" {
		os.Chmod(path, 0o644)
	}
	if err := os.Rename(temporary, path); err != nil {
		os.Remove(temporary)
		return err
	}
	return nil
}
