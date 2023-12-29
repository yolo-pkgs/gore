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
	"mvdan.cc/xurls/v2"
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

func publish(tag string) error {
	slog.Warn("publishing", slog.String("tag", tag))
	return nil
}

func patch() error {
	ctx := context.Background()

	modPath, err := readGoModPath()
	if err != nil {
		return err
	}
	rxRelaxed := xurls.Relaxed()
	if !rxRelaxed.MatchString(modPath) {
		return fmt.Errorf("go module path is not a valid url: %w", err)
	}

	output, err := runCmd(ctx, "git", "rev-list", "--tags")
	if err != nil {
		return fmt.Errorf("fail listing tags: %w", err)
	}
	tagRefs := strings.Fields(output)
	if len(tagRefs) == 0 {
		slog.Info("no tags found")
		return publish("v0.0.1")
	}

	args := []string{"describe", "--tags"}
	args = append(args, tagRefs...)
	fmt.Println(args)
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
	if len(versions) == 0 {
		slog.Info("no valid go version tags found")
		return publish("v0.0.1")
	}
	sort.Sort(version.Collection(versions))
	fmt.Println(versions)

	return nil
}
