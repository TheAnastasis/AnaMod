package desktopentry

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"anamod/internal/atomicfile"
)

func DataHome() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(dir) {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func Install(dataHome, appID, name, program string, icon []byte) error {
	iconPath := filepath.Join(dataHome, "icons", "hicolor", "512x512", "apps", appID+".png")
	if err := writeIfChanged(iconPath, icon); err != nil {
		return err
	}
	entry := "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=" + name + "\n" +
		"Icon=" + iconPath + "\n" +
		"Exec=" + execValue(program) + "\n" +
		"NoDisplay=true\n"
	return writeIfChanged(filepath.Join(dataHome, "applications", appID+".desktop"), []byte(entry))
}

func execValue(program string) string {
	quoted := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`).Replace(program)
	return strings.NewReplacer(`\`, `\\`, `%`, `%%`).Replace(`"` + quoted + `"`)
}

func writeIfChanged(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return nil
	}
	return atomicfile.Write(path, data)
}
