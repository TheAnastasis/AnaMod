package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"github.com/ncruces/zenity"
)

func windowOwner(win fyne.Window) uintptr {
	native, ok := win.(driver.NativeWindow)
	if !ok {
		return 0
	}
	var owner uintptr
	native.RunNative(func(context any) {
		if handle, ok := context.(driver.WindowsWindowContext); ok {
			owner = handle.HWND
		}
	})
	return owner
}

func nativeFolderPicker(request folderRequest) (string, error) {
	path, err := zenity.SelectFile(zenity.Directory(), zenity.Title(request.title),
		zenity.Filename(request.start), zenity.Attach(request.owner))
	if errors.Is(err, zenity.ErrCanceled) {
		return "", nil
	}
	return path, err
}
