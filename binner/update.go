package binner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/yolo-pkgs/cmdgrace"
	"golang.org/x/sync/errgroup"
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

	for _, bin := range b.Bins {
		bin := bin

		g.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			_, err := cmdgrace.Spawn(ctx, exec.Command("go", "install", fmt.Sprintf("%s@latest", bin.Path)))
			_ = bar.Add(1)
			switch {
			case err == nil:
				return nil
			case errors.Is(err, cmdgrace.ErrTimeout):
				if errors.Is(err, cmdgrace.ErrFailToKill) {
					slog.Warn("failed to kill process after timeout", slog.String("pkg", bin.Path), slog.String("err", err.Error()))
				}
				return fmt.Errorf("timeout: %s: %w", bin.Path, err)
			default:
				return fmt.Errorf("error: %s: %w", bin.Path, err)
			}
		})
	}

	return g.Wait()
}
