package ui

import (
	"anamod/internal/datapack"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (p *patcher) build() fyne.CanvasObject {
	p.updateBox = container.NewVBox()
	p.updateBox.Hide()
	p.updateBtn = widget.NewButtonWithIcon("Update AnaMod", theme.DownloadIcon(), nil)
	p.updateBtn.Importance = widget.HighImportance
	p.updateBtn.Hide()

	p.zone = newDropZone(p.browseGame)
	p.linkRow = p.buildLinkRow()
	p.pickScreen = p.buildPickScreen()
	p.changeBtn = widget.NewButtonWithIcon("Change", theme.FolderOpenIcon(), p.changeGame)
	p.folderLabel = widget.NewLabel("")
	p.folderLabel.Wrapping = fyne.TextWrapBreak
	p.statusIcon = widget.NewIcon(nil)
	p.statusIcon.Hide()
	p.statusLabel = wrappedLabel("")
	p.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	p.detailLabel = wrappedLabel("")
	p.detailLabel.Hide()
	gameCard := widget.NewCard("", "", container.NewVBox(
		container.NewBorder(nil, nil, nil, p.changeBtn, p.folderLabel),
		container.NewBorder(nil, nil, p.statusIcon, nil, p.statusLabel),
		p.detailLabel,
	))

	p.optionsBox = container.NewVBox()

	p.rpcs3Label = wrappedLabel("")
	p.rpcs3Choice = widget.NewSelect(nil, p.pickCandidate)
	p.rpcs3Choice.PlaceHolder = "Select RPCS3 Install..."
	p.rpcs3Choice.Hide()
	p.rpcs3Button = widget.NewButtonWithIcon("Select RPCS3 Install", theme.FolderOpenIcon(), p.browseRPCS3)

	p.installBtn = widget.NewButton("Install", p.pressInstall)
	p.installBtn.Importance = widget.HighImportance
	p.uninstallBtn = widget.NewButton("Uninstall", p.pressUninstall)
	p.footer = wrappedLabel("")

	p.rpcs3Box = container.NewVBox(
		heading("RPCS3 (OPTIONAL)"),
		container.NewBorder(nil, nil, nil, container.NewCenter(p.rpcs3Button), p.rpcs3Label),
		p.rpcs3Choice,
	)
	p.showOptions(p.pack.Manifest.Games[0])

	p.content = container.NewVBox(
		gameCard,
		p.optionsBox,
		p.rpcs3Box,
		widget.NewSeparator(),
		p.installBtn,
		p.uninstallBtn,
		p.footer,
	)
	p.scroll = container.NewVScroll(container.NewPadded(p.content))
	p.showPick()
	return container.NewBorder(p.updateBox, nil, nil, nil,
		container.NewStack(p.pickScreen, p.scroll))
}

func (p *patcher) showPick() {
	p.picking = true
	p.scroll.Hide()
	p.pickScreen.Show()
}

func (p *patcher) showGame() {
	p.picking = false
	p.zone.setMessage(nil, "", "", false)
	p.pickScreen.Hide()
	p.scroll.Show()
}

func (p *patcher) showOptions(game datapack.Game) {
	if p.optionsGame == game.ID {
		return
	}
	p.optionsGame = game.ID
	p.optionsBox.RemoveAll()
	p.checks = map[string]*widget.Check{}
	if game.Platform == "xen" {
		p.rpcs3Box.Hide()
		return
	}
	p.optionsBox.Add(heading("RPCS3 PATCHES"))
	for _, option := range game.RPCS3.Options {
		p.addOption(option.ID, option.Label, option.Help)
	}
	p.rpcs3Box.Show()
}

func (p *patcher) addOption(id, label, help string) {
	check := widget.NewCheck(label, nil)
	check.SetChecked(true)
	p.checks[id] = check
	if help == "" {
		p.optionsBox.Add(check)
		return
	}
	button := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		p.showText(label, help)
	})
	button.Importance = widget.LowImportance
	p.optionsBox.Add(container.NewBorder(nil, nil, nil, button, check))
}

func (p *patcher) relayout() {
	for range 3 {
		before := p.content.MinSize()
		p.content.Refresh()
		if p.content.MinSize() == before {
			return
		}
	}
}

func (p *patcher) setStatus(icon fyne.Resource, status, detail string) {
	if p.picking {
		p.zone.setMessage(icon, status, detail, p.busy)
		return
	}
	if icon != nil {
		p.statusIcon.SetResource(icon)
	}
	showIf(p.statusIcon, icon != nil)
	p.statusLabel.SetText(status)
	p.detailLabel.SetText(detail)
	showIf(p.detailLabel, detail != "")
}
