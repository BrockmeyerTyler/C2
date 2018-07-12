package lr0

import (
	"fmt"
	"bytes"
)

var isNonterminal func(int)bool
var getProduction func(int, int)[]int
var getProductions func(int)[][]int
var symName func(int) string
var eofSymbol int

const (
	OpErr    = 0
	OpShift  = 1
	OpReduce = 2
	OpGoto   = 3
	OpAccept = 4
)

// Struct for holding goto information
type closureItemGoto struct {
	sym int
	to  *closureItemSet
}

func (g closureItemGoto) String() string {
	return fmt.Sprintf("{Sym: %v | To: %v}", symName(g.sym), g.to.id)
}

// Holds closures.
type closureItemSet struct {
	id         int
	items      []closureItem
	parentGoto *closureItemGoto
	gotos      []closureItemGoto
}

func (l closureItemSet) Equals(set closureItemSet) bool {
	if len(l.items) != len(set.items) {
		return false
	}
	for i := range set.items {
		if !l.items[i].Equals(set.items[i]) {
			return false
		}
	}
	return true
}

func (l closureItemSet) String() string {
	s := fmt.Sprintf("{ ID: %v\n  Items:", l.id)
	for i := range l.items {
		s += fmt.Sprintf("\n\t%v", l.items[i])
	}
	if l.gotos != nil {
		s += "\n  Gotos:"
		for i := range l.gotos {
			s += fmt.Sprintf("\n\t%v", l.gotos[i])
		}
	}
	return s + "\n}"
}

// One piece of a closure.
type closureItem struct {
	lhs, rhs int
	index    int
}

func (l closureItem) Equals(item closureItem) bool {
	return l.lhs == item.lhs && l.rhs == item.rhs && l.index == item.index
}

func (l closureItem) IsAccept() bool {
	production := getProduction(l.lhs, l.rhs)
	return len(production) != l.index && production[l.index] == eofSymbol
}

func (l closureItem) IsFinished() bool {
	production := getProduction(l.lhs, l.rhs)
	return l.index == len(production)
}

func (l closureItem) String() string {
	var s= fmt.Sprintf("{ lhs: %v, rhs: %v | ", symName(l.lhs), l.rhs)
	for i, sym := range getProduction(l.lhs, l.rhs) {
		if i == l.index {
			s += "@ "
		}
		s += fmt.Sprintf("%v ", symName(sym))
	}
	if l.IsFinished() {
		return s + "@ }"
	}
	return s + "}"
}

type ParseTableEntry struct {
	Data      int
	ReduceRHS int
	Op        byte
}

func (p ParseTableEntry) String() string {
	return fmt.Sprintf("{%v, %v}", p.Data, p.Op)
}

// Generates the parse table using 'pRules' from "vars.go"
func GenerateParseTable(
		fIsNonterminal func(int)bool,
		fGetProduction func(int, int)[]int,
		fGetProductions func(int)[][]int,
		fGetSymName func(int)string,
		augStartSym int,
		endOfFileSym int,
		symCount int) [][]ParseTableEntry {
	isNonterminal = fIsNonterminal
	getProduction = fGetProduction
	getProductions = fGetProductions
	eofSymbol = endOfFileSym
	symName = fGetSymName

	closures := make([]*closureItemSet, 0, 20)
	newClosures := append(
		make([]*closureItemSet, 0, 20),
		&closureItemSet{items: []closureItem{{lhs: augStartSym, rhs: 0}}},
	)

	// Loop over everything in the 'newClosures' set.
	var closureId int
	for {
		NextLoop:
		if closureId >= len(newClosures) {
			break
		}

		closure := newClosures[closureId]
		isFinishedOrAccept := closure.items[0].IsFinished() || closure.items[0].IsAccept()
		closureId++

		// If closure is not a finished one, it must be computed.
		if !isFinishedOrAccept {

			// Compute closure for current set.
			var nonterminalsProcessed = make([]int, 0, 3)
			var itemId int
			for {
				item := closure.items[itemId]
				rule := getProduction(item.lhs, item.rhs)

				// If symbol after parsing index is nonterminal,
				// Add nonterminal's productions to this closure at parsing index 0.
				if isNonterminal(rule[item.index]) {

					// Check to see if the nonterminal has already been processed.
					// if it has been, skip it.
					nonterm := rule[item.index]
					for _, n := range nonterminalsProcessed {
						if n == nonterm {
							goto AlreadyProcessed
						}
					}

					productions := getProductions(nonterm)
					for i := range productions {
						newItem := closureItem{lhs: nonterm, rhs: i}

						// Check that the new closureItem is unique in the closure.
						for j := range closure.items {
							if newItem.Equals(closure.items[j]) {
								goto ItemNotUnique
							}
						}

						// If the closureItem is unique, add it to the closure.
						closure.items = append(closure.items, newItem)
						ItemNotUnique:
					}
					nonterminalsProcessed = append(nonterminalsProcessed, nonterm)
				}
				AlreadyProcessed:

				itemId++
				if itemId >= len(closure.items) {
					break
				}
			}
		}

		// Check to see if the computed closure is equivalent to another existing one.
		// If it is, modify parent's goto to point to the original.
		// After that, drop the current closure and continue on.
		for _, compare := range closures {
			if compare.Equals(*closure) {
				closure.parentGoto.to = compare
				goto NextLoop
			}
		}

		// Closure is not a duplicate.
		closure.id = len(closures)
		closures = append(closures, closure)
		if isFinishedOrAccept {
			continue
		}

		closure.gotos = make([]closureItemGoto, 0, 6)

		for _, item := range closure.items {
			// Create a new closure for each closureItem.
			// Each new closure has the parsing index moved up by one.
			newClosure := &closureItemSet{items: []closureItem{{item.lhs, item.rhs, item.index + 1}}}

			// Set the new closure as a goto destination.
			// Then get a reference to that goto destination.
			// If the new closure is a duplicate, it will need to modify the goto to point at the original.
			production := getProduction(item.lhs, item.rhs)
			closure.gotos = append(closure.gotos, closureItemGoto{production[item.index], newClosure})
			newClosure.parentGoto = &closure.gotos[len(closure.gotos)-1]
			newClosures = append(newClosures, newClosure)
		}
	}
	for c := range closures {
		closures[c].id = c
	}

	/*for c := range closures {
		fmt.Println(closures[c])
	}*/

	errorClosures := make([]*closureItemSet, 0, 5)
	parseTable := make([][]ParseTableEntry, len(closures))
	for i := range closures {
		parseTable[i] = make([]ParseTableEntry, symCount)

		// Check if this is an acceptance state. If so, mark the EOF entry as 'accept'.
		if closures[i].items[0].IsAccept() {
			parseTable[i][eofSymbol].Op = OpAccept

		// Check if this is a final state. If so, mark all pTerminals as 'reduce'.
		} else if closures[i].gotos == nil {
			for e := 0; !isNonterminal(e); e++ {
				parseTable[i][e].Data = closures[i].items[0].lhs
				parseTable[i][e].Op = OpReduce
				parseTable[i][e].ReduceRHS = closures[i].items[0].rhs
			}

			// For each goto, set pTerminals as 'shift' and nonterminals as 'goto'.
		} else {
			for _, g := range closures[i].gotos {
				e := &parseTable[i][g.sym]
				if e.Op != OpErr {
					errorClosures = append(errorClosures, closures[i])
				}
				e.Data = g.to.id
				if isNonterminal(g.sym) {
					e.Op = OpGoto
				} else {
					e.Op = OpShift
				}
			}
		}
	}

	fmt.Println("LR(0) parse table generation complete.")
	for i := range errorClosures {
		fmt.Println("[ParseTableError] Multiple op for symbol in closure.")
		fmt.Println(errorClosures[i])
	}

	return parseTable
}

func BytePrintCode(table [][]ParseTableEntry) []byte {
	var buf bytes.Buffer
	buf.WriteString("[][]LR0ParseTableEntry {\n")
	for _, p := range table {
		buf.WriteString("{")
		for _, s := range p {
			buf.WriteString(fmt.Sprintf("{%v, %v, %v}, ", s.Data, s.ReduceRHS, s.Op))
		}
		buf.WriteString("},\n")
	}
	buf.WriteString("}")
	return buf.Bytes()
}
