package binner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/sync/errgroup"

	"github.com/yolo-pkgs/cmdgrace"

	"github.com/yolo-pkgs/gore/pkg/goproxy"
	"github.com/yolo-pkgs/gore/pkg/gosystem"
)

type Bin struct {
	Binary      string
	Path        string
	Mod         string
	ModVersion  string
	LastVersion string
	Updatable   bool
}

type Binner struct {
	Bins    []Bin
	binPath string
	simple  bool
}

func New(simple bool) (*Binner, error) {
	binPath, err := gosystem.GetBinPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get go bin path: %w", err)
	}

	return &Binner{
		binPath: binPath,
		simple:  simple,
	}, nil
}

func (b *Binner) sortBinsByName() {
	sort.Slice(b.Bins, func(i, j int) bool {
		return b.Bins[i].Binary < b.Bins[j].Binary
	})
}

func (b *Binner) fillBins() error {
	ctx := context.Background()

	output, err := cmdgrace.Spawn(ctx, exec.Command("go", "version", "-m", b.binPath))
	if err != nil {
		return fmt.Errorf("failed to get binaries info: %w", err)
	}

	lines := strings.Split(output, "\n")
	bins := make([]Bin, 0)

	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		var first rune

		for _, element := range line {
			first = element
			break
		}

		if unicode.IsSpace(first) {
			continue
		}

		bin, err := parseBin(b.binPath+"/", line, lines[i+1], lines[i+2])
		if err != nil {
			return fmt.Errorf("failed parsing binary info: %w", err)
		}

		bins = append(bins, bin)
	}

	b.Bins = bins

	return nil
}

func parseBin(binPrefix, binLine, pathLine, modLine string) (Bin, error) {
	fields := strings.Fields(binLine)
	if len(fields) == 0 {
		return Bin{}, errors.New("binLine: len(fields) == 0")
	}
	binNameLong := strings.TrimRight(fields[0], ":")
	binName := strings.TrimPrefix(binNameLong, binPrefix)

	fields = strings.Fields(pathLine)
	if len(fields) < 2 {
		return Bin{}, errors.New("pathLine: len(fields) < 2")
	}
	binPath := fields[1]

	fields = strings.Fields(modLine)
	if len(fields) < 3 {
		return Bin{}, errors.New("modLine: len(fields) < 3")
	}
	mod := fields[1]
	modVersion := fields[2]

	return Bin{
		Binary:     binName,
		Path:       binPath,
		Mod:        mod,
		ModVersion: modVersion,
		Updatable:  false,
	}, nil
}

func (b *Binner) fillProxyUpdateInfo() {
	result := make(chan []Bin)
	binChan := make(chan Bin)

	go func() {
		newBins := make([]Bin, 0)
		for bin := range binChan {
			newBins = append(newBins, bin)
		}
		result <- newBins
		close(result)
	}()

	g := new(errgroup.Group)

	for _, bin := range b.Bins {
		bin := bin

		g.Go(func() error {
			latest, err := goproxy.GetLatestVersion(bin.Mod)
			if err != nil {
				latest = err.Error()
			} else {
				latest = "v" + latest
				// TODO: move this to separate method
				if latest != bin.ModVersion {
					bin.Updatable = true
				}
			}
			bin.LastVersion = latest
			binChan <- bin

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Panic(err)
	}

	close(binChan)

	processedBins := <-result
	b.Bins = processedBins
}
