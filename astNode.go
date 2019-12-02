package c2

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/c2/c2gram"
	"strings"
)

type ASTNode struct {
	symbol     c2gram.Symbol
	up         *ASTNode
	down       []*ASTNode
	production *c2gram.Production
	token      *Token
	file       string
	row        int
	column     int
	// The constructed data for this symbol.
	Data interface{}
}

func (n *ASTNode) Traverse() error {
	hasActions := n.production != nil && n.production.Actions != nil
	if n.down != nil {
		for i, down := range n.down {
			if hasActions {
				if action, ok := n.production.Actions[i]; ok {
					if err := action(n); err != nil {
						return err
					}
				}
			}
			if err := down.Traverse(); err != nil {
				return err
			}
		}
	}
	if hasActions {
		if action, ok := n.production.Actions[len(n.production.RHS)]; ok {
			if err := action(n); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n ASTNode) Symbol() c2gram.Symbol {
	return n.symbol
}

// Go to the symbol above this one in the tree.
// This does not come without risks.
func (n ASTNode) Up() *ASTNode {
	return n.up
}

// For a production, this will contain the symbols that make up the production rule.
func (n ASTNode) Down(i int) *ASTNode {
	return n.down[i]
}

// Get the location of the symbol in code.
func (n ASTNode) Loc() (file string, row, column int) {
	return n.file, n.row, n.column
}

func (n ASTNode) NewError(reason string) error {
	return errors.Errorf("[Error] In %v: At %v:%v | %v", n.file, n.row, n.column, reason)
}

func (n ASTNode) ToString(symbols []c2gram.Symbol) string {
	if n.token != nil {
		return fmt.Sprint("TOKEN: ", n.symbol.Name(), "\tLEXEME: ", n.token.lexeme)
	}
	rhs := make([]string, 0, len(n.production.RHS))
	for _, sym := range n.production.RHS {
		rhs = append(rhs, symbols[sym].Name())
	}
	return fmt.Sprint("SYMBOL:\t", n.symbol.Name(), " ::= ", strings.Join(rhs, " "))
}
