package main

import (
	"fmt"
	c2 "github.com/tjbrockmeyer/C2"
	"github.com/tjbrockmeyer/C2/example/lang"
	"github.com/tjbrockmeyer/C2/lr0"
)

func main() {
	g, err := lang.Build()
	if err != nil {
		panic(err)
	}
	pt, err := lr0.GenerateParseTable(g, true)
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
