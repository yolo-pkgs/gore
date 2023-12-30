package binner

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/yolo-pkgs/gore/pkg/gitversion"
)

func (b *Binner) fillGitUpdateInfo() {
	m := &sync.Mutex{}
	g := new(errgroup.Group)

	for i, bin := range b.Bins {
		i := i
		bin := bin

		if !gitversion.IsGitVersion(bin.ModVersion) {
			continue
		}

		g.Go(func() error {
			gitURL := fmt.Sprintf("https://%s.git", bin.Mod)
			// TODO: edgy...
			if !strings.Contains(bin.Mod, "git") {
				gitURLPre, err := gitversion.FollowRedirect(b.client, fmt.Sprintf("https://%s", bin.Mod))
				if err != nil {
					log.Printf("could not find redirect link for %s: %v\n", bin.Mod, err)
					return nil
				}
				gitURL = fmt.Sprintf("%s.git", gitURLPre)
			}

			commitHash, commitTime, err := gitversion.CloneAndRetrieveLastCommitInfo(gitURL)
			if err != nil {
				log.Printf("failed getting last commit for %s: %v\n", bin.Mod, err)
				return nil
			}

			goHash := commitHash[:12] // TODO: possible panic
			goTimePreprocess := commitTime.UTC().Format("2006 01 02 15 04 05")
			goTimeF := strings.Fields(goTimePreprocess)
			goTime := strings.Join(goTimeF, "")

			goDevVersion := fmt.Sprintf("v0.0.0-%s-%s", goTime, goHash)

			m.Lock()
			defer m.Unlock()
			b.Bins[i].LastVersion = goDevVersion
			return nil
		})

		if err := g.Wait(); err != nil {
			log.Panic("should not panic when resolving git versions")
		}
	}
}
