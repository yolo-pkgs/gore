package version

import (
	"debug/buildinfo"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Binary struct {
	Filename   string
	Path       string
	Mod        string
	ModVersion string
	GoVersion  string
}

type Holder struct {
	Bins []Binary
}

func RunVersion(arg string) ([]Binary, error) {
	_, err := os.Stat(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil, err
	}
	holder := &Holder{
		Bins: make([]Binary, 0),
	}

	if err := scanDir(holder, arg); err != nil {
		return nil, err
	}

	return holder.Bins, nil
}

// scanDir scans a directory for binary to run scanFile on.
func scanDir(holder *Holder, dir string) error {
	binOut := make(chan Binary)
	res := make(chan []Binary)

	go func() {
		bins := make([]Binary, 0)
		for bin := range binOut {
			bins = append(bins, bin)
		}
		res <- bins
		close(res)
	}()

	wg := &sync.WaitGroup{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, _ error) error {
		if d.Type().IsRegular() || d.Type()&fs.ModeSymlink != 0 {
			dir := dir
			path := path
			wg.Add(1)
			go func() {
				defer wg.Done()
				info, err := d.Info()
				if err != nil {
					return
				}
				binary := scanFile(dir, path, info)
				binOut <- binary
			}()
		}

		return nil
	})
	if err != nil {
		return err
	}

	wg.Wait()
	close(binOut)

	all := <-res
	holder.Bins = all

	return nil
}

// scanFile scans file to try to report the Go and module versions.
// If mustPrint is true, scanFile will report any error reading file.
// Otherwise (mustPrint is false, because scanFile is being called
// by scanDir) scanFile prints nothing for non-Go binaries.
func scanFile(arg, file string, info fs.FileInfo) Binary {
	if info.Mode()&fs.ModeSymlink != 0 {
		// Accept file symlinks only.
		i, err := os.Stat(file)
		if err != nil || !i.Mode().IsRegular() {
			return Binary{}
		}
	}

	bi, err := buildinfo.ReadFile(file)
	if err != nil {
		return Binary{}
	}

	binary := Binary{
		Filename:  strings.TrimPrefix(file, arg+"/"),
		GoVersion: bi.GoVersion,
	}

	bi.GoVersion = "" // suppress printing go version again
	mod := bi.String()

	if len(mod) > 0 {
		modString := mod[:len(mod)-1]
		modF := strings.Fields(modString)

		if len(modF) < 5 { //nolint:gomnd // ok
			return binary
		}
		binary.Path = modF[1]
		binary.Mod = modF[3]
		binary.ModVersion = modF[4]

		// fmt.Printf("\t%s\n", strings.ReplaceAll(mod[:len(mod)-1], "\n", "\n\t"))
		return binary
	}

	return binary
}
