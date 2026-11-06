package rpcs3

import (
	"os"
	"path/filepath"
	"strings"
)

func IsRunning() (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(comm)) == "rpcs3" {
			return true, nil
		}
	}
	return false, nil
}
