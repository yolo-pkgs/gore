package grace

import (
	"io"
	"os/exec"

	"golang.org/x/sys/unix"
)

func Spawn(_ chan<- int, cmd *exec.Cmd) (string, error) {
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}
	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// if pidCallback != nil {
	// pidCallback <- cmd.Process.Pid
	// }

	output, err := io.ReadAll(out)
	if err != nil {
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return string(output), nil
}
