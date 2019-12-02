package c2gram

import (
	"github.com/tjbrockmeyer/c2"
	"log"
	"strings"
)

type ProductionAction func(s *c2.ASTNode) error

type NonTerminalDefinition interface {
	RHS(symbolsAndActions ...interface{}) NonTerminalDefinition
}

type nonTerminal struct {
	name        string
	productions []nonTerminalRHS
}

type nonTerminalRHS struct {
	symbols []string
	actions map[int]ProductionAction
}

// Create a new non-terminal for this grammar.
func (g *Grammar) NewNonTerminal(name string) NonTerminalDefinition {
	n := &nonTerminal{
		name:        name,
		productions: make([]nonTerminalRHS, 0, 2),
	}
	g.nonTerminals = append(g.nonTerminals, n)
	g.nonTerminalsByName[n.name] = n
	return n
}

// Defined actions or symbols to be run when processing this symbol in the grammar.
// Symbols may be defined as separate strings, or as a single string with individual symbols separated by spaces.
func (n *nonTerminal) RHS(symbolsAndActions ...interface{}) NonTerminalDefinition {
	symbols := make([]string, 0, len(symbolsAndActions)-1)
	actions := make(map[int]ProductionAction, 2)
	for _, item := range symbolsAndActions {
		switch t := item.(type) {
		case string:
			symbols = append(symbols, strings.Split(t, " ")...)
		case ProductionAction:
			actions[len(symbols)] = t
		case func(*c2.ASTNode) error:
			actions[len(symbols)] = t
		default:
			log.Println(`ignoring bad 'Symbol' or 'Action' found while defining right hand side for`, n.name, `:`, t)
			return n
		}
	}
	n.productions = append(n.productions, nonTerminalRHS{
		symbols: symbols,
		actions: actions,
	})
	return n
}
