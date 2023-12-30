package binner

import (
	"fmt"
	"strings"
)

func (b *Binner) Dump() error {
	if err := b.fillBins(); err != nil {
		return fmt.Errorf("failed to parse binaries: %w", err)
	}
	b.prettyPrintDump()
	return nil
}

func (b *Binner) prettyPrintDump() {
	b.sortBinsByName()

	output := make([]string, 0)
	for _, bin := range b.Bins {
		cmd := fmt.Sprintf("go install %s@latest", bin.Path)
		output = append(output, cmd)
	}
	fmt.Println(strings.Join(output, "\n"))
}
