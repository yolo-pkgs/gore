package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/yolo-pkgs/cmdgrace"
	"golang.org/x/mod/modfile"
)

func runCmd(ctx context.Context, name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	slog.Info("running process", slog.String("cmd", cmd.String()))
	return cmdgrace.Spawn(ctx, cmd)
}

func readGoModPath() (string, error) {
	dat, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("fail reading go.mod file: %w", err)
	}

	modPath := modfile.ModulePath(dat)
	if modPath == "" {
		return "", errors.New("faulty go module path")
	}
	return modPath, nil
}

func patch() error {
	ctx := context.Background()

	// modPath, err := readGoModPath()
	// if err != nil {
	// 	return err
	// }

	output, err := runCmd(ctx, "git", "rev-list", "--tags")
	if err != nil {
		return fmt.Errorf("fail listing tags: %w", err)
	}
	tagRefs := strings.Fields(output)

	args := []string{"describe", "--tags"}
	args = append(args, tagRefs...)
	output, err = runCmd(ctx, "git", args...)
	if err != nil {
		return fmt.Errorf("fail describing tags: %w", err)
	}
	tags := strings.Fields(output)

	versions := make([]*version.Version, 0)
	for _, tag := range tags {
		v, err := version.NewVersion(tag)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}
	sort.Sort(version.Collection(versions))
	fmt.Println(versions)

	return nil
}
