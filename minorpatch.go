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

func publish(modPath, tag string) error {
	ctx := context.Background()
	slog.Info("publishing", slog.String("tag", tag))

	_, err := runCmd(ctx, "git", "tag", tag, "-m", fmt.Sprintf("'fix: %s'", tag))
	if err != nil {
		return fmt.Errorf("failed tagging: %w", err)
	}

	output, err := runCmd(ctx, "git", "push", "origin", fmt.Sprintf("refs/tags/%s", tag))
	if err != nil {
		return fmt.Errorf("failed pushing new tag: %w", err)
	}
	fmt.Println(output)

	output, err = runCmd(ctx, "go", "list", "-m", fmt.Sprintf("%s@%s", modPath, tag))
	if err != nil {
		return fmt.Errorf("failed listing new version: %w", err)
	}

	slog.Info("published new version", slog.String("version", tag))

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

	_, err = runCmd(ctx, "git", "fetch", "--tags")
	if err != nil {
		return fmt.Errorf("fail fetching tags: %w", err)
	}

	output, err := runCmd(ctx, "git", "tag")
	if err != nil {
		return fmt.Errorf("fail getting tags: %w", err)
	}
	tags := strings.Fields(output)
	if len(tags) == 0 {
		slog.Info("no tags found")
		return publish(modPath, "v0.0.1")
	}

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
		return publish(modPath, "v0.0.1")
	}
	sort.Sort(version.Collection(versions))
	lastVersion := versions[len(versions)-1]
	segments := lastVersion.Segments64()
	if len(segments) != 3 {
		return errors.New("number of segments in last version != 3")
	}

	newTag := fmt.Sprintf("v%d.%d.%d", segments[0], segments[1], segments[2]+1)
	return publish(modPath, newTag)
}
