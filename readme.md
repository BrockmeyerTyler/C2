# C2: A Compiler-Compiler
A meta-compiler written in Go that I created so that I can use syntax-highlighted
  go code to create a new language.

# Usage
1. Define `NonTerminals` and `Terminals` using a `Grammar`.
2. Once all non-terminals and terminals are defined, 
  use `Grammar.Assemble()` to get an assembled grammar.
3. Choose a parse table type (currently only an LR(0) is available (it's *basically* an SLR(1)))
  and create the table using the assembled grammar.
4. Create a `c2.Parser` using the parse table and assembled grammar.
5. Read from a file or string to begin parsing your constructed language.
```go
package main
import (
    "fmt"
    "github.com/tjbrockmeyer/c2"
    "github.com/tjbrockmeyer/c2/c2gram"
    "github.com/tjbrockmeyer/c2/c2lr0"
    "strconv"
)

func main() {
    p, err := Build()
    if err != nil {
        panic(err)
    }

    // Should print "3"
    if err := p.ReadString("1 + 2"); err != nil {
        panic(err)
    }
    // Expecting an error here (x is not a valid token)
    if err := p.ReadString("10 + x"); err != nil {
        panic(err)
    }
}

func Build() (*c2.Parser, error) {
    g := c2gram.New()
    g.NewTerminal("num", `[0-9]+`).Action(func(t *c2.Token) (interface{}, error) {
        i, err := strconv.ParseInt(t.Lexeme(), 10, 32)
        return int(i), err
    })
    g.NewNonTerminal("START").
        RHS("EXPR", func(n *c2.ASTNode) error {
            fmt.Println(n.Down(0).Data)
            return nil
        })
    g.NewNonTerminal("EXPR").
        RHS("num + num", func(n *c2.ASTNode) error {
            n.Data = n.Down(0).Data.(int) + n.Down(2).Data.(int)
            return nil
        })
    a, err := g.Assemble()
    if err != nil {
        return nil, err
    }
    
    pt, err := c2lr0.NewParseTable(a, false)
    if err != nil {
        return nil, err
    }
    return c2.NewParser(a, pt), nil
}
```

# TODO
[_] Add serializing stuff for AssembledGrammar and Parser
[_] More docs