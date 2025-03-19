package binner

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/olekukonko/tablewriter"
)

const spinnerMs = 40
const checkingDevMsg = "Checking dev packages for updates..."

func (b *Binner) LSBins() error {
	// Read file metadata including mod path and version
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	b.sortBinsByName()

	names := make([]string, len(b.Bins))

	for i, bin := range b.Bins {
		names[i] = bin.Binary
	}

	color.Cyan(strings.Join(names, "  "))

	return nil
}

func (b *Binner) ListBins() error {
	// Read file metadata including mod path and version
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	// Set private marker on bins
	b.fillPrivateInfo()

	b.StartSpinner("Checking public packages for updates...")
	b.fillProxyUpdateInfo()
	b.StopSpinner()

	// b.spin = spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond)
	b.StartSpinner("Checking private packages for updates...")
	b.fillPrivateUpdateInfo()
	b.StopSpinner()

	// Fetch possible* updates for go dev versions
	if b.checkDev {
		b.StartSpinner(checkingDevMsg)
		b.fillGitUpdateInfo()
		b.StopSpinner()
	}

	// Print out
	b.fillUpdateStatus()
	b.sortBinsByName()
	b.prettyPrintList()

	return nil
}

func (b *Binner) fillUpdateStatus() {
	for i, bin := range b.Bins {
		if b.Bins[i].ModVersion == "(devel)" {
			b.Bins[i].Updatable = true
			continue
		}

		current, err := version.NewVersion(bin.ModVersion)
		if err != nil {
			continue
		}

		last, err := version.NewVersion(bin.LastVersion)
		if err != nil {
			continue
		}

		if current.String() == last.String() {
			b.Bins[i].Updatable = false
			continue
		}

		if last.GreaterThan(current) {
			b.Bins[i].Updatable = true
		}
	}
}

func (b *Binner) prettyPrintList() {
	if b.simple {
		b.simpleOutput()
	} else {
		b.tableOutput()
	}
}

func (b *Binner) simpleOutput() {
	output := make([]string, len(b.Bins))

	for i, bin := range b.Bins {
		updateField := "-"
		if bin.Updatable {
			updateField = bin.LastVersion
		}

		var line string

		if b.extra {
			line = fmt.Sprintf(
				"%s %s %s %s %s %d",
				bin.Binary,
				bin.ModVersion,
				updateField,
				fmt.Sprintf("https://%s", bin.Mod),
				bin.ModTime.Format(time.RFC3339),
				bin.Size,
			)
		} else {
			line = fmt.Sprintf(
				"%s %s %s %s",
				bin.Binary,
				bin.ModVersion,
				updateField,
				fmt.Sprintf("https://%s", bin.Mod),
			)
		}
		output[i] = line
	}

	fmt.Println(strings.Join(output, "\n"))
	fmt.Printf("%s, %d binaries\n", b.binPath, len(b.Bins))
}

func (b *Binner) writeTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)

	if b.extra {
		table.SetHeader([]string{"bin", "version", "update", "module uri", "modified", "size"})
	} else {
		table.SetHeader([]string{"bin", "version", "update", "module uri"})
	}

	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()
}

func (b *Binner) constructDataForTable(bins []Bin) [][]string {
	data := make([][]string, 0)

	for _, bin := range bins {
		updateField := "-"
		if bin.Updatable {
			updateField = bin.LastVersion
		}

		if b.extra {
			data = append(data, []string{
				bin.Binary,
				bin.ModVersion,
				updateField,
				fmt.Sprintf("https://%s", bin.Mod),
				bin.ModTime.Format("Mon, 02 Jan 2006 15:04:05 MST"),
				humanize.Bytes(uint64(bin.Size)),
			})
		} else {
			data = append(data, []string{
				bin.Binary,
				bin.ModVersion,
				updateField,
				fmt.Sprintf("https://%s", bin.Mod),
			})
		}
	}

	return data
}

func (b *Binner) tableOutput() {
	if b.group {
		b.grouppedTableOutput()
		return
	}

	data := b.constructDataForTable(b.Bins)

	b.writeTable(data)
	color.Cyan(fmt.Sprintf("%s, %d binaries\n", b.binPath, len(b.Bins)))
}

func (b *Binner) grouppedTableOutput() {
	m := make(map[string][]Bin)

	for _, bin := range b.Bins {
		if len(bin.Mod) == 0 {
			log.Printf("module path not set on output, please report")
			continue
		}
		domain := strings.Split(bin.Mod, "/")[0]
		_, ok := m[domain]

		if !ok {
			m[domain] = []Bin{bin}
		} else {
			bins := m[domain]
			bins = append(bins, bin)
			m[domain] = bins
		}
	}

	for domain, bins := range m {
		data := b.constructDataForTable(bins)
		b.writeTable(data)
		color.Cyan(fmt.Sprintf("%s, %d binaries\n", domain, len(bins)))
		fmt.Println()
		fmt.Println()
	}
}
