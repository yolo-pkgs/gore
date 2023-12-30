package gosystem

import (
	"errors"
	"fmt"
	"os"
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
