package ui

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (p *patcher) showText(title, text string) {
	label := wrappedLabel(text)
	d := dialog.NewCustom(title, "OK", label, p.win)
	d.Resize(fyne.NewSize(460, 0))
	d.Show()
}

func (p *patcher) confirm(title, text, action string, then func()) {
	label := wrappedLabel(text)
	d := dialog.NewCustomConfirm(title, action, "Cancel", label, func(ok bool) {
		if ok {
			then()
		}
	}, p.win)
	d.Resize(fyne.NewSize(460, 0))
	d.Show()
}

type playerError struct {
	title   string
	message string
	cause   error
}

func (e *playerError) Error() string {
	return e.message + " (" + e.cause.Error() + ")"
}

func (e *playerError) Unwrap() error {
	return e.cause
}

func (p *patcher) showError(title, message string, err error) {
	if told, ok := errors.AsType[*playerError](err); ok {
		if told.title != "" {
			title = told.title
		}
		message, err = told.message, told.cause
	}
	content := container.NewVBox(wrappedLabel(message))
	if err != nil {
		fmt.Fprintf(p.log, "Error: %s: %s (%v)\n", title, message, err)
	} else {
		fmt.Fprintf(p.log, "Error: %s: %s\n", title, message)
	}
	copied := widget.NewLabel("")
	copyLog := widget.NewButtonWithIcon("Copy Log", theme.ContentCopyIcon(), func() {
		p.app.Clipboard().SetContent(p.log.Text())
		copied.SetText("Copied! Paste it into your bug report.")
	})
	openFolder := widget.NewButtonWithIcon("Open Log Folder", theme.FolderOpenIcon(),
		p.openLogFolder)
	content.Add(container.NewHBox(copyLog, openFolder))
	content.Add(copied)
	d := dialog.NewCustom(title, "Close", content, p.win)
	d.Resize(fyne.NewSize(480, 0))
	d.Show()
}

func (p *patcher) openLogFolder() {
	if p.logDir == "" {
		return
	}
	folder, err := url.Parse(storage.NewFileURI(p.logDir).String())
	if err == nil {
		err = p.app.OpenURL(folder)
	}
	if err != nil {
		p.showText("Log Folder", "The log is in "+p.logDir)
	}
}

func wrappedLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func showIf(object fyne.CanvasObject, visible bool) {
	if visible {
		object.Show()
	} else {
		object.Hide()
	}
}

func heading(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

func sentence(err error) string {
	text := err.Error()
	if text == "" {
		return text
	}
	runes := []rune(text)
	runes[0] = unicode.ToUpper(runes[0])
	text = string(runes)
	if !strings.ContainsRune(".?!", runes[len(runes)-1]) {
		text += "."
	}
	return text
}
