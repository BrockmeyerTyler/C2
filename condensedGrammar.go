package c2

import (
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

type Symbol interface {
	ID() int
	IsNonTerminal() bool
	Name() string
}

type Terminal interface {
	Symbol
	Find([]byte) []byte
	IsIgnored() bool
	RunAction(*Token) (interface{}, error)
}

type NonTerminal interface {
	Symbol
}

type Production struct {
	ID      int
	LHS     int
	RHS     []int
	Actions map[int]ProductionAction
}

type CondensedGrammar struct {
	Undefined        int
	EndOfFile        int
	AugmentedStart   int
	Symbols          []Symbol
	Productions      []*Production
	ProductionsByLHS map[int][]*Production
}

func (g Grammar) Build() (CondensedGrammar, error) {
	symbols := make([]Symbol, 0, 40)
	productions := make([]*Production, 40)
	productionsByLHS := make(map[int][]*Production, 20)
	idsByName := make(map[string]int, 40)

	undefined := &terminal{
		name:  "*Undefined",
		regex: regexp.MustCompile(`.`),
	}
	endOfFile := &terminal{
		name:  "*EndOfFile",
		regex: regexp.MustCompile(`^$`),
	}
	augStart := &nonTerminal{
		name: "*AugmentedStart",
		productions: []nonTerminalRHS{{
			symbols: []string{g.nonTerminals[0].name, endOfFile.name},
			actions: map[int]ProductionAction{},
		}},
	}
	terminals := append(append(append(make([]*terminal, 0, len(g.terminals)+2), g.terminals...), undefined), endOfFile)
	nonTerminals := append(append(make([]*nonTerminal, 0, len(g.nonTerminals)+1), augStart), g.nonTerminals...)

	// Add terminals and nonterminals to the list of symbols, and add productions to their lookup lists
	undefinedSymbols := make([]string, 0, 20)
	for id, t := range terminals {
		idsByName[t.name] = id
		symbols = append(symbols, condensedSymbol{
			id:       id,
			terminal: t,
		})
	}
	for _, n := range nonTerminals {
		id := len(symbols)
		idsByName[n.name] = id
		productionsByLHS[id] = make([]*Production, 0, len(n.productions))
		for _, rhs := range n.productions {
			p := &Production{
				ID:      len(productions),
				LHS:     id,
				RHS:     make([]int, 0, len(rhs.symbols)),
				Actions: rhs.actions,
			}
			productions = append(productions, p)
			productionsByLHS[id] = append(productionsByLHS[id], p)
		}
		symbols = append(symbols, condensedSymbol{
			id:          id,
			nonTerminal: n,
		})
	}

	for _, n := range nonTerminals {
		for rhsId, rhs := range n.productions {
			p := productionsByLHS[idsByName[n.name]][rhsId]
			for _, s := range rhs.symbols {
				if id, ok := idsByName[s]; ok {
					p.RHS = append(p.RHS, id)
				} else {
					undefinedSymbols = append(undefinedSymbols, s)
				}
			}
		}
	}

	if len(undefinedSymbols) > 0 {
		return CondensedGrammar{}, errors.Errorf(
			"the following symbols are undefined in the grammar: %s", strings.Join(undefinedSymbols, ", "))
	}
	return CondensedGrammar{
		Undefined:        len(terminals) - 2,
		EndOfFile:        len(terminals) - 1,
		AugmentedStart:   len(terminals),
		Symbols:          symbols,
		Productions:      productions,
		ProductionsByLHS: productionsByLHS,
	}, nil
}

type condensedSymbol struct {
	id int
	*terminal
	*nonTerminal
}

func (s condensedSymbol) ID() int {
	return s.id
}

func (s condensedSymbol) IsNonTerminal() bool {
	return s.nonTerminal != nil
}

func (s condensedSymbol) Name() string {
	if s.IsNonTerminal() {
		return s.nonTerminal.name
	}
	return s.terminal.name
}
