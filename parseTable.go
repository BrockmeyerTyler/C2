package c2

import (
	"fmt"
	"strings"
)

type ParseTableEntry struct {
	Data int
	Op   byte
}

type ParseTable [][]ParseTableEntry

func (p ParseTable) ToString(symbols []Symbol) string {
	stateSpacing := 1
	k := len(p)
	for k > 10 {
		stateSpacing++
		k /= 10
	}
	stateSpacing++
	b := strings.Builder{}
	b.WriteString(fmt.Sprint("      ", strings.Repeat(" ", stateSpacing), " | "))
	for _, s := range symbols {
		spacing := len(s.Name())
		if spacing < 6 {
			spacing = 6
		}
		b.WriteString(fmt.Sprint(s.Name(), strings.Repeat(" ", spacing-len(s.Name())), " | "))
	}
	b.WriteString("\n")
	for i, row := range p {
		b.WriteString(fmt.Sprint("state ", i, strings.Repeat(" ", stateSpacing-len(fmt.Sprint(i))), " | "))
		for j, entry := range row {
			spacing := len(symbols[j].Name())
			if spacing < 6 {
				spacing = 6
			}
			b.WriteString(fmt.Sprint(opNames[entry.Op], " ", fmt.Sprint(entry.Data), strings.Repeat(" ", spacing-4-len(fmt.Sprint(entry.Data))), " | "))
		}
		b.WriteString("\n")
	}
	return b.String()
}
