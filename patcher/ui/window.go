package ui

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"anamod/internal/datapack"
	"anamod/internal/desktopentry"
	"anamod/internal/install"
	"anamod/internal/logfile"
	"anamod/internal/rpcs3"
	"anamod/internal/update"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	windowWidth  = 560
	windowHeight = 540
)

const noWindowMessage = "AnaMod Patcher couldn't open due to this machine not having an OpenGL 2.1 driver. " +
	"Please update the graphics driver and try again."

type widgets struct {
	content      *fyne.Container
	scroll       *container.Scroll
	zone         *dropZone
	linkRow      *fyne.Container
	pickScreen   fyne.CanvasObject
	changeBtn    *widget.Button
	updateBox    *fyne.Container
	updateBtn    *widget.Button
	statusIcon   *widget.Icon
	folderLabel  *widget.Label
	statusLabel  *widget.Label
	detailLabel  *widget.Label
	checks       map[string]*widget.Check
	optionsBox   *fyne.Container
	rpcs3Box     *fyne.Container
	rpcs3Label   *widget.Label
	rpcs3Choice  *widget.Select
	rpcs3Button  *widget.Button
	installBtn   *widget.Button
	uninstallBtn *widget.Button
	footer       *widget.Label
}

type patcher struct {
	widgets

	app      fyne.App
	alert    func(string)
	openURL  func(*url.URL) error
	win      fyne.Window
	log      *logfile.Log
	logDir   string
	userDir  string
	version  string
	repo     string
	pack     *datapack.Pack
	settings settings

	root   string
	result *install.Result
	loc    *rpcs3.Location
	busy   bool
	work   sync.WaitGroup

	nativeFolder func(folderRequest) (string, error)
	pickWait     *widget.PopUp

	picking     bool
	optionsGame string
}

func Run(app fyne.App, embedded []byte, version, repo string) {
	program, err := os.Executable()
	logDir := ""
	if err == nil {
		logDir = filepath.Dir(program)
	}
	p := newPatcher(app, embedded, version, repo, logDir)
	defer p.log.Close()
	if runtime.GOOS == "linux" && err == nil {
		p.installDesktopEntry(program)
	}
	p.win.SetCloseIntercept(p.closeRequested)
	p.win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			p.chooseFolder(uris[0].Path())
		}
	})
	app.Lifecycle().SetOnStarted(p.started)
	go p.checkForUpdate()
	p.win.ShowAndRun()
}

func (p *patcher) started() {
	p.stopIfNoWindow()
	if hasWindow(p.win) {
		p.restoreGame()
	}
}

func (p *patcher) stopIfNoWindow() {
	if hasWindow(p.win) {
		return
	}
	fmt.Fprintln(p.log, noWindowMessage)
	p.alert(noWindowMessage)
	p.app.Quit()
}

func hasWindow(win fyne.Window) bool {
	native, ok := win.(driver.NativeWindow)
	if !ok {
		return true
	}
	opened := true
	native.RunNative(func(context any) {
		switch handle := context.(type) {
		case driver.WindowsWindowContext:
			opened = handle.HWND != 0
		case driver.X11WindowContext:
			opened = handle.WindowHandle != 0
		case driver.WaylandWindowContext:
			opened = handle.WaylandSurface != 0
		}
	})
	return opened
}

func newPatcher(app fyne.App, embedded []byte, version, repo, logDir string) *patcher {
	p := &patcher{app: app, alert: alertUser, openURL: app.OpenURL,
		nativeFolder: nativeFolderPicker, version: version, repo: repo}
	p.win = app.NewWindow("AnaMod Patcher " + version)
	p.win.SetFixedSize(true)
	p.openLog(logDir)
	log.SetFlags(0)
	log.SetOutput(io.MultiWriter(p.log, os.Stderr))
	fmt.Fprintf(p.log, "AnaMod Patcher %s on %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
	var downloaded []byte
	if logDir != "" {
		p.userDir = logDir
		p.settings = loadSettings(logDir)
		downloaded, _ = os.ReadFile(filepath.Join(logDir, update.DataAsset))
	} else {
		fmt.Fprintln(p.log, "No folder for settings: the patcher's own folder is unknown")
	}
	pack, err := pickPack(embedded, downloaded, version)
	if err != nil {
		fmt.Fprintf(p.log, "Error: %v\n", err)
		label := wrappedLabel(sentence(err) + " Download the patcher again.")
		p.win.SetContent(container.NewPadded(label))
		p.win.Resize(fyne.NewSize(480, 200))
		return p
	}
	p.pack = pack
	fmt.Fprintf(p.log, "Using data for AnaMod %s\n", pack.Manifest.AnamodVersion)
	p.win.SetContent(p.build())
	p.win.Resize(fyne.NewSize(windowWidth, windowHeight))
	p.findRPCS3()
	p.refresh()
	return p
}

func (p *patcher) installDesktopEntry(program string) {
	icon := p.app.Icon()
	if icon == nil {
		return
	}
	dataHome, err := desktopentry.DataHome()
	if err == nil {
		err = desktopentry.Install(dataHome, p.app.UniqueID(), "AnaMod Patcher", program, icon.Content())
	}
	if err != nil {
		fmt.Fprintf(p.log, "No desktop icon: %v\n", err)
	}
}

func (p *patcher) openLog(dir string) {
	for _, dir := range []string{dir, os.TempDir()} {
		if dir == "" {
			continue
		}
		if log, err := logfile.Open(dir); err == nil {
			p.log, p.logDir = log, dir
			return
		}
	}
	p.log = logfile.Memory()
}

func (p *patcher) changeGame() {
	if p.busy {
		return
	}
	p.result = nil
	p.root = ""
	p.showPick()
	p.refresh()
}

func (p *patcher) isXen() bool {
	return p.result != nil && p.result.Game.Platform == "xen"
}

func (p *patcher) refresh() {
	state := install.NoGame
	if p.result != nil {
		state = p.result.State
	}
	canInstall, updating := p.installFlags(state)
	p.refreshChecks(canInstall)
	p.refreshInstallButton(state, canInstall, updating)
	p.refreshUninstallButton()
	p.refreshFooter(state)
	p.refreshRPCS3Line()
	p.relayout()
}

func (p *patcher) installFlags(state install.State) (canInstall, updating bool) {
	canInstall = !p.busy && (state == install.Vanilla || state == install.Installed)
	updating = state == install.Installed && p.packIsNewer(p.result.Record)
	return canInstall, updating
}

func (p *patcher) refreshChecks(canInstall bool) {
	for _, check := range p.checks {
		if canInstall && (p.loc != nil || p.isXen()) {
			check.Enable()
		} else {
			check.Disable()
		}
	}
}

func (p *patcher) refreshInstallButton(state install.State, canInstall, updating bool) {
	switch {
	case updating:
		p.installBtn.SetText("Update to AnaMod " + p.pack.Manifest.AnamodVersion)
	case state == install.Installed:
		p.installBtn.SetText("Update")
	default:
		p.installBtn.SetText("Install")
	}
	if canInstall {
		p.installBtn.Enable()
	} else {
		p.installBtn.Disable()
	}
}

func (p *patcher) refreshUninstallButton() {
	hasRecord := p.result != nil && p.result.Record != nil
	if hasRecord {
		p.uninstallBtn.Show()
	} else {
		p.uninstallBtn.Hide()
	}
	if hasRecord && !p.busy {
		p.uninstallBtn.Enable()
	} else {
		p.uninstallBtn.Disable()
	}
}

func (p *patcher) refreshFooter(state install.State) {
	if state == install.ChangedSinceInstall {
		p.footer.SetText("You can still uninstall. It restores the original scripts, so " +
			"changes made to them after AnaMod was installed will be lost!")
		return
	}
	p.footer.SetText("")
}

func (p *patcher) refreshRPCS3Line() {
	if p.loc != nil {
		p.rpcs3Label.SetText("Found at " + p.loc.Folder)
		p.rpcs3Button.SetText("Change")
	} else if !p.rpcs3Choice.Visible() {
		p.rpcs3Label.SetText("Not found... You only need this for the RPCS3 patches.")
		p.rpcs3Button.SetText("Where's RPCS3?")
	}
}

func (p *patcher) packIsNewer(record *install.Record) bool {
	return datapack.CompareVersions(p.pack.Manifest.AnamodVersion, record.AnamodVersion) > 0
}

func (p *patcher) background(message string, work func(), done func()) {
	p.busy = true
	p.setStatus(theme.ViewRefreshIcon(), message, "")
	p.refresh()
	p.work.Go(func() {
		work()
		fyne.Do(func() {
			p.busy = false
			done()
			p.refresh()
		})
	})
}

func (p *patcher) browseGame() {
	if p.busy {
		return
	}
	p.pickFolder(folderRequest{title: "Choose Your THP8 Folder"}, func(path string) {
		if path != "" {
			p.chooseFolder(path)
		}
	})
}

func (p *patcher) chooseFolder(path string) {
	if p.busy || p.pack == nil || p.pickWait != nil {
		return
	}
	fmt.Fprintf(p.log, "Folder chosen: %s\n", path)
	p.showPick()
	p.result = nil
	root, err := install.FindRoot(path)
	if err != nil {
		fmt.Fprintf(p.log, "%v\n", err)
		p.root = ""
		p.setStatus(theme.ErrorIcon(), sentence(err), path)
		p.refresh()
		return
	}
	p.root = root
	p.folderLabel.SetText(root)
	p.recheck()
}

func (p *patcher) chosenOptions() []string {
	if p.result == nil || (!p.isXen() && p.loc == nil) {
		return nil
	}
	var chosen []string
	for _, id := range p.result.Game.OptionIDs() {
		if check := p.checks[id]; check != nil && check.Checked {
			chosen = append(chosen, id)
		}
	}
	return chosen
}

func (p *patcher) pressInstall() {
	if p.busy || p.result == nil {
		return
	}
	if p.loc != nil && !p.isXen() && p.rpcs3IsOpen() {
		return
	}
	p.runInstall()
}

func (p *patcher) runInstall() {
	root, result, pack, loc, chosen := p.root, *p.result, p.pack, p.loc, p.chosenOptions()
	if p.isXen() {
		loc = nil
	}
	var summary string
	var err error
	p.background("Installing AnaMod...", func() {
		summary, err = doInstall(root, result, pack, loc, chosen, p.log)
	}, func() {
		if err != nil {
			p.showError("AnaMod Failed to Install", "Something went wrong while installing... "+
				"The window now shows what state your game files are in. If you report this, "+
				"please include the log!!!", err)
		} else {
			fmt.Fprintf(p.log, "Done:\n%s\n", summary)
			p.showText("Done", summary)
		}
		p.recheck()
	})
}

func (p *patcher) pressUninstall() {
	if p.busy || p.result == nil || p.result.Record == nil {
		return
	}
	message := "This reverts THP8's original scripts and unticks AnaMod patches in RPCS3."
	if p.isXen() {
		message = "This puts THP8's original scripts back."
	}
	if p.result.State == install.ChangedSinceInstall {
		message += " Changes made to these scripts after AnaMod was installed will be lost."
	}
	root := p.root
	p.confirm("Uninstall AnaMod?", message, "Uninstall", func() { p.runUninstall(root) })
}

func (p *patcher) runUninstall(root string) {
	if p.busy || p.result == nil || p.result.Record == nil || p.root != root {
		fmt.Fprintln(p.log, "Uninstall cancelled... the patcher was busy or the game folder changed.")
		return
	}
	record := p.result.Record
	var loc *rpcs3.Location
	if record.RPCS3Dir != "" {
		found, err := rpcs3.FromFolder(record.RPCS3Dir, runtime.GOOS)
		if err == nil {
			loc = &found
		} else {
			fmt.Fprintf(p.log, "RPCS3 folder (%s) no longer works: %v\n",
				record.RPCS3Dir, err)
		}
	}
	if loc != nil && p.rpcs3IsOpen() {
		return
	}
	result := *p.result
	var err error
	p.background("Uninstalling AnaMod...", func() {
		err = doUninstall(root, result, loc, p.log)
	}, func() {
		if err != nil {
			p.showError("Uninstall Failed", "Something went wrong while uninstalling... "+
				"The window now shows what state your game files are in. If you report this, "+
				"please include the log!!!", err)
		} else {
			p.showText("Done", "AnaMod is uninstalled!")
		}
		p.recheck()
	})
}

func (p *patcher) checkForUpdate() {
	if p.repo == "" || p.pack == nil {
		fmt.Fprintln(p.log, "Update Check Skipped...")
		return
	}
	client := &http.Client{Timeout: 30 * time.Second}
	release, err := update.Latest(client, fmt.Sprintf(update.APIURLTemplate, p.repo))
	if err != nil {
		fmt.Fprintf(p.log, "Update Check Failed: %v\n", err)
		return
	}
	current := p.pack.Manifest.AnamodVersion
	fmt.Fprintf(p.log, "Latest Release: AnaMod %s (using %s)\n", release.Version, current)
	if datapack.CompareVersions(release.Version, current) <= 0 {
		return
	}
	fyne.Do(func() { p.offerUpdate(client, release) })
}

func (p *patcher) offerUpdate(client *http.Client, release update.Release) {
	p.updateBtn.OnTapped = func() { p.showChangelog(client, release) }
	p.updateBtn.Enable()
	p.updateBtn.Show()
	p.linkRow.Refresh()
}

func (p *patcher) showChangelog(client *http.Client, release update.Release) {
	body := strings.TrimSpace(release.Body)
	if body == "" {
		body = "This release has no notes."
	}
	notes := widget.NewRichTextFromMarkdown(body)
	notes.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(notes)
	scroll.SetMinSize(fyne.NewSize(420, 220))
	current := p.pack.Manifest.AnamodVersion
	content := container.NewBorder(
		wrappedLabel("You're using AnaMod "+current+". Changelog:"),
		nil, nil, nil, scroll,
	)
	d := dialog.NewCustomConfirm("AnaMod "+release.Version+" Is Available!", "Update", "Cancel",
		content, func(ok bool) {
			if ok {
				p.updateBtn.Disable()
				go p.downloadUpdate(client, release)
			}
		}, p.win)
	d.Resize(fyne.NewSize(480, 380))
	d.Show()
}

func (p *patcher) downloadUpdate(client *http.Client, release update.Release) {
	data, err := update.Download(client, release.DataURL)
	var pack *datapack.Pack
	if err == nil {
		pack, err = openUpdate(data, p.version, p.pack)
	}
	fyne.Do(func() {
		switch {
		case errors.Is(err, datapack.ErrNeedsNewerPatcher):
			fmt.Fprintf(p.log, "AnaMod %s needs a newer patcher\n", release.Version)
			page, _ := url.Parse(release.PageURL)
			open := widget.NewButtonWithIcon("Open the Download Page", theme.ComputerIcon(),
				func() { p.app.OpenURL(page) })
			p.updateBtn.Hide()
			p.linkRow.Refresh()
			p.showUpdate("AnaMod "+release.Version+" needs a newer patcher than this one.", open)
		case err != nil:
			p.updateBtn.Enable()
			p.showError("Update Download Failed", "The update couldn't be downloaded. "+
				"Press Update AnaMod to try again.", err)
		default:
			if p.userDir != "" {
				if err := os.WriteFile(filepath.Join(p.userDir, update.DataAsset), data,
					0o644); err != nil {
					fmt.Fprintf(p.log, "Could not keep the downloaded data: %v\n", err)
				}
			}
			p.pack = pack
			p.updateBtn.Hide()
			p.linkRow.Refresh()
			version := pack.Manifest.AnamodVersion
			fmt.Fprintf(p.log, "Now using data for AnaMod %s\n", version)
			p.showUpdate("Now using AnaMod "+version+".", nil)
			if !p.busy {
				p.recheck()
			}
		}
	})
}

func (p *patcher) showUpdate(message string, action fyne.CanvasObject) {
	p.updateBox.RemoveAll()
	row := container.NewHBox(widget.NewIcon(theme.InfoIcon()), widget.NewLabel(message))
	if action != nil {
		row.Add(action)
	}
	p.updateBox.Add(row)
	p.updateBox.Show()
}

func (p *patcher) closeRequested() {
	if !p.busy {
		p.win.Close()
		return
	}
	p.confirm("AnaMod Is Still Working",
		"AnaMod is still working on your game files. Quitting now can leave them in "+
			"a bad state.", "Quit Anyway", p.win.Close)
}
