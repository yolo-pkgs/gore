package binner

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"

	"github.com/yolo-pkgs/gore/pkg/gitversion"
	"github.com/yolo-pkgs/gore/pkg/grace"
)

func (b *Binner) Update() error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	if b.checkDev {
		b.m.Lock()
		b.spin = spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond)
		b.spin.Suffix = checkingDev
		b.spin.Start()
		b.m.Unlock()
		b.fillGitUpdateInfo()
		b.Sstop()
	}

	b.fillUpdateStatus()

	if err := b.update(); err != nil {
		return fmt.Errorf("some updates failed: %w", err)
	}

	color.Cyan("All binaries updated!")

	return nil
}

func (b *Binner) update() error {
	g := new(errgroup.Group)
	bar := progressbar.NewOptions(
		len(b.Bins),
		progressbar.OptionSetDescription("updating"),
		progressbar.OptionSetWidth(35),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	limit := runtime.NumCPU() * doubleCPU
	limiter := make(chan struct{}, limit)

	for i := 0; i < limit; i++ {
		limiter <- struct{}{}
	}

	for _, bin := range b.Bins {
		bin := bin

		<-limiter

		g.Go(func() error {
			_, err := grace.Spawn(nil, exec.Command("go", "install", fmt.Sprintf("%s@latest", bin.Path)))
			_ = bar.Add(1)
			if err != nil {
				limiter <- struct{}{}
				return fmt.Errorf("error: %s: %w", bin.Path, err)
			}

			// NOTE: somehow they are not installed on @latest command
			if bin.Updatable && gitversion.IsGitVersion(bin.LastVersion) {
				pkgIdentifier := fmt.Sprintf("%s@%s", bin.Path, bin.LastVersion)
				if _, err = grace.Spawn(nil, exec.Command("go", "list", "-m", pkgIdentifier)); err != nil {
					log.Printf("failed go list -m on devel package %s: %v\n", bin.Path, err)
				}
				if _, err = grace.Spawn(nil, exec.Command("go", "install", pkgIdentifier)); err != nil {
					log.Printf("failed go install on devel package %s: %v\n", bin.Path, err)
				}
			}

			limiter <- struct{}{}

			return nil
		})
	}

	return g.Wait()
}
