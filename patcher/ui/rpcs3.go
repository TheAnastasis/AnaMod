package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"anamod/internal/rpcs3"
)

func (p *patcher) findRPCS3() {
	if p.settings.RPCS3Folder != "" {
		loc, err := rpcs3.FromFolder(p.settings.RPCS3Folder, runtime.GOOS)
		if err == nil {
			p.loc = &loc
			fmt.Fprintf(p.log, "RPCS3: %s (remembered)\n", loc.Folder)
			return
		}
		fmt.Fprintf(p.log, "The remembered RPCS3 folder no longer works? %v\n", err)
	}
	home, _ := os.UserHomeDir()
	candidates := rpcs3.Candidates(runtime.GOOS, home, os.Getenv("XDG_CONFIG_HOME"))
	fmt.Fprintf(p.log, "RPCS3 folders found: %v\n", candidates)
	switch len(candidates) {
	case 0:
	case 1:
		p.useRPCS3(candidates[0])
	default:
		p.rpcs3Label.SetText("Multiple RPCS3 installs detected. Which one " +
			"do you use?")
		p.rpcs3Choice.Options = candidates
		p.rpcs3Choice.Show()
	}
}

func (p *patcher) pickCandidate(folder string) {
	if folder != "" {
		p.useRPCS3(folder)
	}
}

func (p *patcher) useRPCS3(folder string) {
	loc, err := rpcs3.FromFolder(folder, runtime.GOOS)
	if err != nil {
		fmt.Fprintf(p.log, "RPCS3 folder %s: %v\n", folder, err)
		p.showError("That Isn't RPCS3's Folder", sentence(err), nil)
		return
	}
	p.loc = &loc
	p.rpcs3Choice.Hide()
	fmt.Fprintf(p.log, "RPCS3: %s (hard drive %s)\n", loc.Folder, loc.DevHDD0)
	p.settings.RPCS3Folder = loc.Folder
	p.keepSettings()
	p.refresh()
}

func (p *patcher) browseRPCS3() {
	if p.busy {
		return
	}
	request := folderRequest{title: "Choose RPCS3 Install Location"}
	if runtime.GOOS != "windows" {
		request.title = "Choose RPCS3's Settings Folder (With config.yml)"
		if home, err := os.UserHomeDir(); err == nil {
			request.start = filepath.Join(home, ".config")
		}
	}
	p.pickFolder(request, func(path string) {
		if path != "" {
			p.useRPCS3(path)
		}
	})
}

func (p *patcher) rpcs3IsOpen() bool {
	running, err := rpcs3.IsRunning()
	if err != nil {
		fmt.Fprintf(p.log, "Could not tell whether RPCS3 is running: %v\n", err)
	}
	if running {
		fmt.Fprintln(p.log, "RPCS3 is running...")
		p.showText("Close RPCS3 First!", "RPCS3 rewrites its settings when it closes, so "+
			"AnaMod can't set it up while it is open. Close RPCS3, then try again!")
	}
	return running
}
