package binner

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/briandowns/spinner"
	"github.com/go-resty/resty/v2"

	"github.com/yolo-pkgs/gore/pkg/goproxy"
	"github.com/yolo-pkgs/gore/pkg/gosystem"
	"github.com/yolo-pkgs/gore/pkg/modversion"
)

type Bin struct {
	Binary      string
	Path        string
	Mod         string
	ModVersion  string
	LastVersion string
	Updatable   bool
	Private     bool
	Size        int64
	ModTime     time.Time
}

type Binner struct {
	m            *sync.Mutex
	spin         *spinner.Spinner
	client       *resty.Client
	Bins         []Bin
	binPath      string
	simple       bool
	checkDev     bool
	extra        bool
	group        bool
	privateGlobs []string
}

func (b *Binner) StartSpinner(msg string) {
	b.m.Lock()
	b.spin.Suffix = " " + msg
	b.spin.Start()
	b.m.Unlock()
}

func (b *Binner) StopSpinner() {
	b.m.Lock()
	b.spin.Stop()
	b.m.Unlock()
}

func New(simple, checkDev, extra, group bool) (*Binner, error) {
	binPath, err := gosystem.GetBinPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get go bin path: %w", err)
	}

	privateGlobs, err := gosystem.GoPrivate()
	if err != nil {
		return nil, fmt.Errorf("failed to get privateGlobs: %w", err)
	}

	client := resty.New()

	binner := &Binner{
		m:            &sync.Mutex{},
		spin:         spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond),
		client:       client,
		binPath:      binPath,
		simple:       simple,
		checkDev:     checkDev,
		privateGlobs: privateGlobs,
		extra:        extra,
		group:        group,
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		<-sigc
		// NOTE: need to stop spinner before exiting - it messes up the shell.
		binner.m.Lock()
		if binner.spin.Enabled() {
			binner.spin.Stop()
		}

		os.Exit(1)
	}()

	return binner, nil
}

func (b *Binner) sortBinsByName() {
	sort.Slice(b.Bins, func(i, j int) bool {
		return b.Bins[i].Binary < b.Bins[j].Binary
	})
}

func (b *Binner) fillBins() error {
	bins, err := modversion.RunVersion(b.binPath)
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
			Private:    false,
			Size:       bin.Size,
			ModTime:    bin.ModTime,
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

	limit := runtime.NumCPU()
	limiter := make(chan struct{}, limit)

	for range limit {
		limiter <- struct{}{}
	}

	for _, bin := range b.Bins {
		bin := bin

		<-limiter

		g.Go(func() error {
			if bin.Private {
				binChan <- bin
				limiter <- struct{}{}
				return nil
			}

			latest, err := goproxy.GetLatestVersion(b.client, bin.Mod)
			if err != nil {
				latest = err.Error()
			} else {
				latest = "v" + latest
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
