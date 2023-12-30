package binner

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
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
	t := table.NewWriter()

	fmt.Printf("%s, %d binaries\n", b.binPath, len(b.Bins))

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
	} else {
		t.AppendHeader(table.Row{"bin", "package", "version", "update"})
		b.sortBinsByName()
		for _, bin := range b.Bins {

			updateField := "-"
			if bin.Updatable {
				updateField = bin.LastVersion
			}

			t.AppendRow(table.Row{
				bin.Binary,
				fmt.Sprintf("https://%s", bin.Path),
				bin.ModVersion,
				updateField,
			})
		}
		fmt.Println(t.Render())
	}
}
