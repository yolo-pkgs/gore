package binner

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/yolo-pkgs/gore/pkg/grace"
)

func (b *Binner) IsPrivate(bin Bin) bool {
	for _, glob := range b.privateGlobs {
		match, err := path.Match(strings.TrimSpace(glob), bin.Mod)
		if err != nil {
			// TODO: for now assume it's public.
			log.Printf("error matching private glob %s: %v\n", glob, err)
			continue
		}

		if match {
			return true
		}
	}

	return false
}

func (b *Binner) fillPrivateInfo() {
	for i, bin := range b.Bins {
		if b.IsPrivate(bin) {
			b.Bins[i].Private = true
		}
	}
}

func (b *Binner) fillPrivateUpdateInfo() {
	for i, bin := range b.Bins {
		if !bin.Private {
			continue
		}

		// git ls-remote --heads --tags --refs --sort=-committerdate https://github.com/yolo-pkgs/gore.git
		// NOTE: 'committerdate' requires access to object data
		//
		// 4ab6401800c54ce363fc943ac2b89f9b2e97cce6        refs/heads/main
		// 1ce4b0a43b703e0b5b01dc5b64d8df38ecec60fd        refs/tags/v0.0.1
		output, err := grace.Spawn(nil, exec.Command(
			"git", "ls-remote", "--heads", "--tags", "--refs",
			fmt.Sprintf("https://%s.git", bin.Mod),
		))
		if err != nil {
			log.Printf("failed to fetch git remote tags for repo %s: %v", bin.Mod, err)
			continue
		}

		versions := make([]*version.Version, 0)

		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}
			lineF := strings.Fields(line)
			fullTag := lineF[1]
			fullTagS := strings.Split(fullTag, "/")
			tag := fullTagS[2]

			v, err := version.NewVersion(tag)
			if err != nil {
				continue
			}

			versions = append(versions, v)
		}

		if len(versions) == 0 {
			continue
		}

		sort.Sort(version.Collection(versions))
		lastVersion := versions[len(versions)-1]

		b.Bins[i].LastVersion = lastVersion.String()
	}
}
