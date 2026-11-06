package rpcs3

import (
	"os/exec"
	"strings"
	"syscall"
)

func IsRunning() (bool, error) {
	command := exec.Command("tasklist", "/FO", "CSV", "/NH", "/FI", "IMAGENAME eq rpcs3.exe")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	output, err := command.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToLower(string(output)), `"rpcs3.exe"`), nil
}
