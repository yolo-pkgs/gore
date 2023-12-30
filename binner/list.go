package binner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

const spinnerMs = 50

func (b *Binner) ListBins() error {
	// Read file metadata including mod path and version
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	// Set private marker on bins
	b.fillPrivateInfo()

	// Fetch updates for public bins
	spin := spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond)
	spin.Suffix = " Checking public packages for updates..."
	spin.Start()
	b.fillProxyUpdateInfo()
	spin.Stop()

	// Fetch updates for private bins
	spin = spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond)
	spin.Suffix = " Checking private packages for updates..."
	spin.Start()
	b.fillPrivateUpdateInfo()
	spin.Stop()

	// Fetch possible* updates for go dev versions
	if b.checkDev {
		spin = spinner.New(spinner.CharSets[14], spinnerMs*time.Millisecond)
		spin.Suffix = " Checking dev packages for updates..."
		spin.Start()
		b.fillGitUpdateInfo()
		spin.Stop()
	}

	// Print out
	b.sortBinsByName()
	b.prettyPrintList()

	return nil
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
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
				bin.ModTime.Format(time.RFC3339),
				bin.Size,
			)
		} else {
			line = fmt.Sprintf(
				"%s %s %s %s",
				bin.Binary,
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
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
		table.SetHeader([]string{"bin", "package", "version", "update", "modified", "size"})
	} else {
		table.SetHeader([]string{"bin", "package", "version", "update"})
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
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
				bin.ModTime.Format("Mon, 02 Jan 2006 15:04:05 MST"),
				humanize.Bytes(uint64(bin.Size)),
			})
		} else {
			data = append(data, []string{
				bin.Binary,
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
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
