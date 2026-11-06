package ui

import (
	"syscall"
	"unsafe"
)

const messageBoxIconError = 0x10

func alertUser(text string) {
	title, _ := syscall.UTF16PtrFromString("AnaMod Patcher")
	body, _ := syscall.UTF16PtrFromString(text)
	syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(body)),
		uintptr(unsafe.Pointer(title)), messageBoxIconError)
}
