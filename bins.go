package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"unicode"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/yolo-pkgs/cmdgrace"
	"golang.org/x/sync/errgroup"
)

const goProxyURL = "https://proxy.golang.org"

type Bin struct {
	Binary      string
	Path        string
	Mod         string
	ModVersion  string
	LastVersion string
	Updatable   string
}

func getLatestVersion(moduleName string) (string, error) {
	url := fmt.Sprintf("%s/%s/@v/list", goProxyURL, moduleName)
	client := resty.New()

	resp, err := client.R().
		Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("http: %d", resp.StatusCode())
	}

	lines := strings.Fields(string(resp.Body()))
	if len(lines) == 0 {
		return "", errors.New("")
	}

	versions := make([]*version.Version, 0)
	for _, tag := range lines {
		v, err := version.NewVersion(tag)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(version.Collection(versions))
	lastVersion := versions[len(versions)-1]

	return lastVersion.String(), nil
}

func getBinPath() (string, error) {
	gobinEnv := os.Getenv("GOBIN")
	if gobinEnv != "" {
		return gobinEnv, nil
	}

	gopathEnv := os.Getenv("GOPATH")
	if gopathEnv != "" {
		return fmt.Sprintf("%s/bin", gopathEnv), nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("HOME env var not set")
	}

	fallback := fmt.Sprintf("%s/go/bin", home)

	return fallback, nil
}

func postProcessBins(bins []Bin) []Bin {
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
	for _, bin := range bins {
		bin := bin
		g.Go(func() error {
			latest, err := getLatestVersion(bin.Mod)
			if err != nil {
				latest = err.Error()
			} else {
				latest = "v" + latest
				if latest != bin.ModVersion {
					bin.Updatable = "yes"
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
	return processedBins
}

func prettyPrintBins(binPath string, bins []Bin) {
	t := table.NewWriter()
	t.SetTitle(binPath)

	t.AppendHeader(table.Row{"bin", "package", "version", "latest", "update"})
	for _, bin := range bins {
		t.AppendRow(table.Row{bin.Binary, bin.Path, bin.ModVersion, bin.LastVersion, bin.Updatable})
	}

	fmt.Println(t.Render())
}

func listBins() error {
	ctx := context.Background()

	binPath, err := getBinPath()
	if err != nil {
		return fmt.Errorf("failed to get go bin path: %w", err)
	}

	// output, err := runCmd(ctx, "go", "version", "-m", binPath)
	output, err := cmdgrace.Spawn(ctx, exec.Command("go", "version", "-m", binPath))
	if err != nil {
		return fmt.Errorf("failed to get binaries info: %w", err)
	}

	lines := strings.Split(output, "\n")
	bins := make([]Bin, 0)

	for i, line := range lines {
		// TODO: probably useless check.
		if len(line) == 0 {
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

		bin, err := parseBin(binPath+"/", line, lines[i+1], lines[i+2])
		if err != nil {
			return fmt.Errorf("failed parsing binary info: %w", err)
		}
		bins = append(bins, bin)
	}

	bins = postProcessBins(bins)
	prettyPrintBins(binPath, bins)
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
	}, nil
}
