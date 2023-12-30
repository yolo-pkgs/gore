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
	m *sync.Mutex
	Bins []Binary
}

func RunVersion(arg string) []Binary {
	_, err := os.Stat(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}
	holder := &Holder{
		m: &sync.Mutex{},
		Bins: make([]Binary, 0),
	}
	scanDir(holder, arg)
	return holder.Bins
}

// scanDir scans a directory for binary to run scanFile on.
func scanDir(holder *Holder, dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, _ error) error {
		if d.Type().IsRegular() || d.Type()&fs.ModeSymlink != 0 {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			binary := scanFile(dir, path, info)
			holder.m.Lock()
			defer holder.m.Unlock()
			holder.Bins = append(holder.Bins, binary)
		}
		return nil
	})
}

// isGoBinaryCandidate reports whether the file is a candidate to be a Go binary.
func isGoBinaryCandidate(file string, info fs.FileInfo) bool {
	if info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
		return true
	}
	name := strings.ToLower(file)
	switch filepath.Ext(name) {
	case ".so", ".exe", ".dll":
		return true
	default:
		return strings.Contains(name, ".so.")
	}
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
		if len(modF) < 5 {
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
