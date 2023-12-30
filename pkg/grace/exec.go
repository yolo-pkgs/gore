package grace

import (
	"io"
	"os/exec"
)

func Spawn(_ chan<- int, cmd *exec.Cmd) (string, error) {
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err = cmd.Start(); err != nil {
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
