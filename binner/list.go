package binner

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func (b *Binner) ListBins() error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	b.fillProxyUpdateInfo()
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
			line := fmt.Sprintf("%s %s %s %s", bin.Binary, fmt.Sprintf("https://%s", bin.Path), bin.ModVersion, updateField)
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

			data = append(data, []string{
				bin.Binary,
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"bin", "package", "version", "update"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.AppendBulk(data)
		caption := fmt.Sprintf("%s, %d binaries\n", b.binPath, len(b.Bins))
		table.SetCaption(true, caption)
		table.Render()
	}
}
