package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const settingsFile = "anamod.ini"

type settings struct {
	RPCS3Folder string
	GameFolder  string
}

func settingsPath(userDir string) string {
	return filepath.Join(userDir, settingsFile)
}

func loadSettings(userDir string) settings {
	var s settings
	file, err := os.Open(settingsPath(userDir))
	if err != nil {
		return s
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == ';' || line[0] == '#' || line[0] == '[' {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "rpcs3_folder":
			s.RPCS3Folder = value
		case "game_folder":
			s.GameFolder = value
		}
	}
	return s
}

func saveSettings(userDir string, s settings) error {
	text := "[folders]\nrpcs3_folder = " + s.RPCS3Folder + "\ngame_folder = " + s.GameFolder + "\n"
	return os.WriteFile(settingsPath(userDir), []byte(text), 0o644)
}

func (p *patcher) keepSettings() {
	if p.userDir == "" {
		return
	}
	if err := saveSettings(p.userDir, p.settings); err != nil {
		fmt.Fprintf(p.log, "Could not save settings: %v\n", err)
	}
}

func (p *patcher) rememberGame() {
	if p.root == p.settings.GameFolder {
		return
	}
	p.settings.GameFolder = p.root
	p.keepSettings()
}

func (p *patcher) restoreGame() {
	folder := p.settings.GameFolder
	if folder == "" {
		return
	}
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		fmt.Fprintf(p.log, "The remembered game folder is gone: %s\n", folder)
		return
	}
	fmt.Fprintf(p.log, "Game folder: %s (remembered)\n", folder)
	p.chooseFolder(folder)
}
