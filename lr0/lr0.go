package lr0

import (
	"errors"
	"fmt"
	c2 "github.com/tjbrockmeyer/C2"
	"strings"
)

type _Closure struct {
	id          int
	key         string
	gotos       map[int]*_Closure
	items       []_ClosureItem
	uniqueItems map[string]struct{}
}

// Create a new closure with some items added on creation.
func newClosure(items ..._ClosureItem) *_Closure {
	c := &_Closure{
		items:       make([]_ClosureItem, 0, 8),
		uniqueItems: make(map[string]struct{}, 8),
	}
	b := strings.Builder{}
	for _, i := range items {
		if c.tryAddUniqueItem(i) {
			b.WriteString(fmt.Sprint(i.production.ID, ".", i.index))
			b.WriteString("|")
		}
	}
	c.key = b.String()
	return c
}

// Add a unique item to the closure. Return true on success, false otherwise.
func (c *_Closure) tryAddUniqueItem(i _ClosureItem) bool {
	if _, ok := c.uniqueItems[i.asKey()]; ok {
		return false
	}
	c.uniqueItems[i.asKey()] = struct{}{}
	c.items = append(c.items, i)
	return true
}

type _ClosureItem struct {
	production *c2.Production
	index      int
}

func (c _ClosureItem) isReducer() bool {
	symbols := c.production.RHS
	return c.index == len(symbols)
}

func (c _ClosureItem) asKey() string {
	return fmt.Sprint(c.production.ID, "|", c.index)
}

func itemIsAccepter(c _ClosureItem, startSymbolProduction *c2.Production) bool {
	return c.production.ID == startSymbolProduction.ID && c.index == len(startSymbolProduction.RHS)-1
}

func closureToString(c *_Closure, syms []c2.Symbol) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprint("(", c.id, ")"))
	for _, item := range c.items {
		b.WriteString(fmt.Sprint("\n\t", syms[item.production.LHS].Name(), " ->"))
		symbols := item.production.RHS
		for i, symbol := range symbols {
			if item.index == i {
				b.WriteString(" @")
			}
			b.WriteString(" " + syms[symbol].Name())
		}
		if item.index == len(symbols) {
			b.WriteString(" @")
		}
	}
	for symbol, gotoClosure := range c.gotos {
		b.WriteString(fmt.Sprint("\n\ton ", syms[symbol].Name(), " goto ", gotoClosure.id))
	}
	return b.String()
}

func GenerateParseTable(g c2.CondensedGrammar, ignoreShiftReduce bool) (c2.ParseTable, error) {
	startProduction := g.ProductionsByLHS[g.AugmentedStart][0]
	// Create original closure.
	c := newClosure(_ClosureItem{production: startProduction})
	closuresByKey := make(map[string]*_Closure, 20)
	closuresByKey[c.key] = c
	closuresById := append(make([]*_Closure, 0, 20), c)
	var closureId int
	for {
		if closureId == len(closuresById) {
			break
		}
		closure := closuresById[closureId]
		closureId++

		// Loop over all items in the closure:
		for i := 0; i < len(closure.items); i++ {
			item := closure.items[i]
			rhs := item.production.RHS
			if item.index == len(rhs) {
				continue
			}
			symbol := rhs[item.index]

			// If the current item has a non-terminal next, create items for that non-terminal's productions at index 0.
			if symbol >= g.AugmentedStart {
				for _, p := range g.ProductionsByLHS[symbol] {
					closure.tryAddUniqueItem(_ClosureItem{
						production: p,
						index:      0,
					})
				}
			}
		}

		// For each item in the closure,
		// If it is not a reducer or accepter,
		// Mark down the symbol and create an item for which to start with in the Goto closure.
		// If there are multiple items to be created using the same symbol, they will be added to a list.
		// Create an additional list for symbols to ensure a deterministic outcome (avoid looping over a map)
		nextClosureItems := make(map[int][]_ClosureItem, 8)
		nextClosureSymbols := make([]int, 0, 8)
		for _, item := range closure.items {
			rhs := item.production.RHS
			if itemIsAccepter(item, startProduction) || item.isReducer() {
				continue
			}
			symbol := rhs[item.index]
			if _, ok := nextClosureItems[symbol]; !ok {
				nextClosureItems[symbol] = make([]_ClosureItem, 0, 2)
				nextClosureSymbols = append(nextClosureSymbols, symbol)
			}
			nextClosureItems[symbol] = append(nextClosureItems[symbol], _ClosureItem{
				production: item.production,
				index:      item.index + 1,
			})
		}

		// Create a new closure for every unique symbol following the items' indices.
		// If the closure matches an existing one, set the goto for that symbol to the existing closure.
		// Otherwise, create the closure as a valid new closure.
		closure.gotos = make(map[int]*_Closure, 8)
		for _, symbol := range nextClosureSymbols {
			c := newClosure(nextClosureItems[symbol]...)
			if matchedClosure, ok := closuresByKey[c.key]; ok {
				closure.gotos[symbol] = matchedClosure
				continue
			}
			c.id = len(closuresById)
			closuresByKey[c.key] = c
			closuresById = append(closuresById, c)
			closure.gotos[symbol] = c
		}
	}

	// Build the parse table
	parseTable := make(c2.ParseTable, len(closuresById))
	for state := 0; state < len(closuresById); state++ {
		closure := closuresById[state]
		parseTable[state] = make([]c2.ParseTableEntry, len(g.Symbols))

		// Set up accept op for this state.
		if itemIsAccepter(closure.items[0], startProduction) {
			parseTable[state][g.EndOfFile].Op = c2.OpAccept
			continue
		}

		// Set up reduce ops for this state.
		for _, item := range closure.items {
			if item.isReducer() {
				for t := 0; t < g.AugmentedStart; t++ {
					parseTable[state][t].Op = c2.OpReduce
					parseTable[state][t].Data = item.production.ID
				}
			}
		}

		// Set up shift ops for this state.
		for symbol, toClosure := range closure.gotos {
			entry := &parseTable[state][symbol]
			entry.Data = toClosure.id
			if symbol >= g.AugmentedStart {
				entry.Op = c2.OpGoto
			} else {
				entry.Op = c2.OpShift
			}
		}
	}

	reduceReduceConflicts := make(map[int]*_Closure, 5)
	shiftReduceConflicts := make(map[int]*_Closure, 5)
	for _, closure := range closuresById {
		var hasShift bool
		var hasReduce bool
		for _, item := range closure.items {
			if itemIsAccepter(item, startProduction) {
				continue
			}
			if item.isReducer() {
				hasReduce = true
				if hasShift {
					shiftReduceConflicts[closure.id] = closure
				}
				continue
			}
			hasShift = true
			if hasReduce {
				shiftReduceConflicts[closure.id] = closure
			}
		}
	}

	b := strings.Builder{}
	for _, closure := range reduceReduceConflicts {
		b.WriteString(fmt.Sprintln("reduce/reduce conflict", closureToString(closure, g.Symbols)))
	}
	if !ignoreShiftReduce {
		for _, closure := range shiftReduceConflicts {
			b.WriteString(fmt.Sprintln("shift/reduce conflict", closureToString(closure, g.Symbols)))
		}
	}

	// for _, c := range closuresById {
	// 	fmt.Println(closureToString(c, g.Symbols))
	// }

	if b.Len() > 0 {
		return parseTable, errors.New(b.String())
	}
	return parseTable, nil
}
