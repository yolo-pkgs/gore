package grace

import (
	"context"
	"os/exec"
)

func Spawn(_ context.Context, cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
