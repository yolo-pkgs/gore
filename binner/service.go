package binner

import (
	"fmt"
	"log"
	"runtime"
	"sort"

	"golang.org/x/sync/errgroup"

	"github.com/yolo-pkgs/gore/pkg/goproxy"
	"github.com/yolo-pkgs/gore/pkg/gosystem"
	"github.com/yolo-pkgs/gore/pkg/version"
)

const doubleCPU = 2

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
	bins, err := version.RunVersion(b.binPath)
	if err != nil {
		return err
	}
	clean := make([]Bin, 0)

	for _, bin := range bins {
		// TODO: support go binaries without module.
		if bin.Mod == "" {
			continue
		}

		clean = append(clean, Bin{
			Binary:     bin.Filename,
			Path:       bin.Path,
			Mod:        bin.Mod,
			ModVersion: bin.ModVersion,
			Updatable:  false,
		})
	}
	b.Bins = clean

	return nil
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

	limit := runtime.NumCPU() * doubleCPU
	limiter := make(chan struct{}, limit)

	for i := 0; i < limit; i++ {
		limiter <- struct{}{}
	}

	for _, bin := range b.Bins {
		bin := bin

		<-limiter

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

			limiter <- struct{}{}

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
