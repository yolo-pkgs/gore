package binner

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func (b *Binner) ListBins() error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}
	b.fillUpdateInfo()
	b.prettyPrintList()
	return nil
}

func (b *Binner) prettyPrintList() {
	t := table.NewWriter()
	// t.SetTitle(fmt.Sprintf("%s    %d binaries", b.binPath, len(b.Bins)))

	fmt.Printf("%s, %d binaries\n", b.binPath, len(b.Bins))

	t.AppendHeader(table.Row{"bin", "package", "version", "latest", "update"})

	b.sortBinsByName()
	for _, bin := range b.Bins {
		t.AppendRow(table.Row{bin.Binary, bin.Path, bin.ModVersion, bin.LastVersion, bin.Updatable})
	}

	fmt.Println(t.Render())
}
