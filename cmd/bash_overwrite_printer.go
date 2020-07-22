package cmd

import (
	"fmt"
	"strings"
)

type BashOverwritePrinter struct {
	row int
}

func (p *BashOverwritePrinter) Print(out string) {
	if p.row > 0 {
		out = "\033[" + fmt.Sprint(p.row) + "A\033[0;K" + out
	}
	row := strings.Count(out, "\n")
	if p.row > row {
		out = strings.Repeat("\033[\n", p.row-row) + out
	} else {
		p.row = row
	}
	fmt.Print(out)
}
