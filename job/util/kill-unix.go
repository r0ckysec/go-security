//go:build !windows
// +build !windows

/**
 * @Description
 * @Author r0cky
 * @Date 2021/10/19 9:26
 **/
package system

import (
	"os/exec"
	"syscall"
)

func SetPgid(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return
}

func GetPPids(pid int) ([]int, error) {
	return []int{}, nil
}

func Kill(pids []uint32) {
	for _, pid := range pids {
		syscall.Kill(int(pid), syscall.SIGKILL)
	}
}

func KillAll(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}
