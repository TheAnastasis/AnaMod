package ui

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/godbus/dbus/v5"
)

const (
	portalService   = "org.freedesktop.portal.Desktop"
	portalObject    = "/org/freedesktop/portal/desktop"
	portalOpenFile  = "org.freedesktop.portal.FileChooser.OpenFile"
	portalRequest   = "org.freedesktop.portal.Request"
	portalResponse  = "Response"
	portalPicked    = uint32(0)
	portalCancelled = uint32(1)
)

func windowOwner(win fyne.Window) uintptr {
	return 0
}

func nativeFolderPicker(request folderRequest) (string, error) {
	bus, err := dbus.SessionBus()
	if err != nil {
		return "", err
	}
	match := []dbus.MatchOption{
		dbus.WithMatchInterface(portalRequest),
		dbus.WithMatchMember(portalResponse),
	}
	if err := bus.AddMatchSignal(match...); err != nil {
		return "", err
	}
	defer bus.RemoveMatchSignal(match...)
	replies := make(chan *dbus.Signal, 8)
	bus.Signal(replies)
	defer bus.RemoveSignal(replies)

	var handle dbus.ObjectPath
	portal := bus.Object(portalService, portalObject)
	call := portal.Call(portalOpenFile, 0, "", request.title, portalOptions(request))
	if err := call.Store(&handle); err != nil {
		return "", err
	}
	for reply := range replies {
		if reply.Path == handle {
			return folderFromReply(reply)
		}
	}
	return "", errors.New("the connection to the portal closed")
}

func portalOptions(request folderRequest) map[string]dbus.Variant {
	options := map[string]dbus.Variant{"directory": dbus.MakeVariant(true)}
	if request.confirm != "" {
		options["accept_label"] = dbus.MakeVariant(request.confirm)
	}
	if request.start != "" {
		options["current_folder"] = dbus.MakeVariant(append([]byte(request.start), 0))
	}
	return options
}

func folderFromReply(reply *dbus.Signal) (string, error) {
	var status uint32
	var results map[string]dbus.Variant
	if err := dbus.Store(reply.Body, &status, &results); err != nil {
		return "", err
	}
	if status == portalCancelled {
		return "", nil
	}
	if status != portalPicked {
		return "", fmt.Errorf("the portal ended the request with status %d", status)
	}
	uris, _ := results["uris"].Value().([]string)
	if len(uris) == 0 {
		return "", nil
	}
	folder, err := url.Parse(uris[0])
	if err != nil {
		return "", err
	}
	if folder.Scheme != "file" {
		return "", fmt.Errorf("%s is not a local folder", uris[0])
	}
	return filepath.Clean(folder.Path), nil
}
