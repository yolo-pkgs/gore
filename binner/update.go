package binner

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"

	"github.com/yolo-pkgs/gore/pkg/grace"
)

func (b *Binner) Update() error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	if err := b.update(); err != nil {
		return fmt.Errorf("some updates failed: %w", err)
	}

	fmt.Println("All binaries updated!")

	return nil
}

func (b *Binner) update() error {
	g := new(errgroup.Group)
	bar := progressbar.Default(int64(len(b.Bins)), "updating")

	limit := runtime.NumCPU() * 2
	limiter := make(chan struct{}, limit)

	for i := 0; i < limit; i++ {
		limiter <- struct{}{}
	}

	for _, bin := range b.Bins {
		bin := bin

		<-limiter

		g.Go(func() error {
			_, err := grace.Spawn(context.Background(), exec.Command("go", "install", fmt.Sprintf("%s@latest", bin.Path)))
			_ = bar.Add(1)
			if err != nil {
				limiter <- struct{}{}
				return fmt.Errorf("error: %s: %w", bin.Path, err)
			}
			limiter <- struct{}{}

			return nil
		})
	}

	return g.Wait()
}
