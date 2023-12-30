package binner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dustin/go-humanize"
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
	b.prettyPrintList()

	return nil
}

func (b *Binner) prettyPrintList() {
	if b.simple {
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
	} else {
		b.sortBinsByName()

		data := make([][]string, 0)

		for _, bin := range b.Bins {
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

		table := tablewriter.NewWriter(os.Stdout)
		if b.extra {
			table.SetHeader([]string{"bin", "package", "version", "update", "modified", "size"})
		} else {
			table.SetHeader([]string{"bin", "package", "version", "update"})
		}
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.AppendBulk(data)
		caption := fmt.Sprintf("%s, %d binaries\n", b.binPath, len(b.Bins))
		table.SetCaption(true, caption)
		table.Render()
	}
}
