package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type folderRequest struct {
	title   string
	confirm string
	start   string
	owner   uintptr
}

func (p *patcher) pickFolder(request folderRequest, done func(path string)) {
	request.owner = windowOwner(p.win)
	wait := p.blockWindow(request, done)
	p.work.Go(func() {
		path, err := p.nativeFolder(request)
		fyne.Do(func() {
			if p.pickWait != wait {
				return
			}
			p.unblockWindow()
			if err != nil {
				fmt.Fprintf(p.log, "No native folder picker: %v\n", err)
				p.showBuiltInPicker(request, done)
				return
			}
			done(path)
		})
	})
}

func (p *patcher) blockWindow(request folderRequest, done func(path string)) *widget.PopUp {
	builtIn := widget.NewButton("Use Patcher's File Browser", func() {
		p.unblockWindow()
		p.showBuiltInPicker(request, done)
	})
	message := widget.NewLabel("Select Folder")
	content := container.NewPadded(container.NewVBox(message, builtIn))
	p.pickWait = widget.NewModalPopUp(content, p.win.Canvas())
	p.pickWait.Show()
	return p.pickWait
}

func (p *patcher) unblockWindow() {
	p.pickWait.Hide()
	p.pickWait = nil
}

func (p *patcher) showBuiltInPicker(request folderRequest, done func(path string)) {
	picker := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil || folder == nil {
			done("")
			return
		}
		done(folder.Path())
	}, p.win)
	picker.SetTitleText(request.title)
	if request.confirm != "" {
		picker.SetConfirmText(request.confirm)
	}
	if request.start != "" {
		start, err := storage.ListerForURI(storage.NewFileURI(request.start))
		if err == nil {
			picker.SetLocation(start)
		}
	}
	picker.Resize(fyne.NewSize(540, 480))
	picker.Show()
}
