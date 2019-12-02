package main

import (
	"fmt"
	c2 "github.com/tjbrockmeyer/c2"
	"github.com/tjbrockmeyer/c2/c2lr0"
	"github.com/tjbrockmeyer/c2/example/lang"
)

func main() {
	g, err := lang.Build()
	if err != nil {
		panic(err)
	}
	pt, err := c2lr0.NewParseTable(g, true)
	if err != nil {
		panic(err)
	}
	// fmt.Println(pt.ToString(g.Symbols))

	p := c2.NewParser(g, pt)
	p.OnNextParse = func(n *c2.ASTNode, nextAction byte) {
		if n.Symbol().IsNonTerminal() {
			fmt.Println(n.ToString(g.Symbols))
		}
	}
	if err = p.ReadFile("./test.lang"); err != nil {
		panic(err)
	}
}
