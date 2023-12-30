package gosystem

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yolo-pkgs/gore/pkg/grace"
)

func GetBinPath() (string, error) {
	gobinEnv := os.Getenv("GOBIN")
	if gobinEnv != "" {
		return gobinEnv, nil
	}

	gopathEnv := os.Getenv("GOPATH")
	if gopathEnv != "" {
		return fmt.Sprintf("%s/bin", gopathEnv), nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("HOME env var not set")
	}

	fallback := fmt.Sprintf("%s/go/bin", home)

	return fallback, nil
}

func GoPrivate() ([]string, error) {
	output, err := grace.Spawn(nil, exec.Command("go", "env", "GOPRIVATE"))
	if err != nil {
		return nil, err
	}

	return strings.Split(output, ","), nil
}
