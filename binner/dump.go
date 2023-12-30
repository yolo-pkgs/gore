package binner

import (
	"fmt"
	"strings"
)

func (b *Binner) Dump(latest bool) error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}

	b.prettyPrintDump(latest)

	return nil
}

func (b *Binner) prettyPrintDump(latest bool) {
	b.sortBinsByName()

	output := make([]string, 0)

	for _, bin := range b.Bins {
		modVersion := bin.ModVersion
		if latest {
			modVersion = "latest"
		}
		cmd := fmt.Sprintf("go install %s@%s", bin.Path, modVersion)
		output = append(output, cmd)
	}
	fmt.Println(strings.Join(output, "\n"))
}
