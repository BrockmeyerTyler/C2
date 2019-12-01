package lang

import (
	"fmt"
	c2 "github.com/tjbrockmeyer/C2"
	"strconv"
)

const (
	vTypeUndefined = iota
	vTypeBool
	vTypeInteger
	vTypeFloat
	vTypeString

	opAdd
	opSub
	opMul
	opDiv
)

var typeToString = []string{
	"bool",
	"int",
	"float",
	"string",
}

var opToString = []string{
	"+",
	"-",
	"*",
	"/",
}

type Value struct {
	value interface{}
	vType int
}

type opHandlerRef struct {
	op, lhs, rhs int
}

var opHandlers = map[opHandlerRef]func(s *c2.ASTNode, lhs, rhs interface{}){
	{opAdd, vTypeBool, vTypeBool}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(bool) || rhs.(bool), vType: vTypeBool}
	},
	{opAdd, vTypeInteger, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(int) + rhs.(int), vType: vTypeInteger}
	},
	{opAdd, vTypeFloat, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(float64) + rhs.(float64), vType: vTypeFloat}
	},
	{opAdd, vTypeInteger, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: float64(lhs.(int)) + rhs.(float64), vType: vTypeFloat}
	},
	{opAdd, vTypeFloat, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: rhs.(float64) + float64(lhs.(int)), vType: vTypeFloat}
	},
	{opAdd, vTypeString, vTypeString}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: rhs.(string) + lhs.(string), vType: vTypeString}
	},

	{opMul, vTypeBool, vTypeBool}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(bool) && rhs.(bool), vType: vTypeBool}
	},
	{opMul, vTypeInteger, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(int) * rhs.(int), vType: vTypeInteger}
	},
	{opMul, vTypeFloat, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(float64) * rhs.(float64), vType: vTypeFloat}
	},
	{opMul, vTypeInteger, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: float64(lhs.(int)) * rhs.(float64), vType: vTypeFloat}
	},
	{opMul, vTypeFloat, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: rhs.(float64) * float64(lhs.(int)), vType: vTypeFloat}
	},

	{opSub, vTypeInteger, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(int) - rhs.(int), vType: vTypeInteger}
	},
	{opSub, vTypeFloat, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(float64) - rhs.(float64), vType: vTypeFloat}
	},
	{opSub, vTypeInteger, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: float64(lhs.(int)) - rhs.(float64), vType: vTypeFloat}
	},
	{opSub, vTypeFloat, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: rhs.(float64) - float64(lhs.(int)), vType: vTypeFloat}
	},

	{opDiv, vTypeInteger, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(int) / rhs.(int), vType: vTypeInteger}
	},
	{opDiv, vTypeFloat, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: lhs.(float64) / rhs.(float64), vType: vTypeFloat}
	},
	{opDiv, vTypeInteger, vTypeFloat}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: float64(lhs.(int)) / rhs.(float64), vType: vTypeFloat}
	},
	{opDiv, vTypeFloat, vTypeInteger}: func(s *c2.ASTNode, lhs, rhs interface{}) {
		s.Data = Value{value: rhs.(float64) / float64(lhs.(int)), vType: vTypeFloat}
	},
}

func Build() (c2.CondensedGrammar, error) {
	variables := make(map[string]*Value)

	g := c2.NewGrammar()
	g.NewTerminal("ws", `\s+`).Ignore()
	g.NewTerminal("nl", `\r?\n`).Ignore()
	g.NewTerminal(";", ";")
	g.NewTerminal("!", "!")
	g.NewTerminal("=", "=")
	g.NewTerminal("(", `\(`)
	g.NewTerminal(")", `\)`)
	g.NewTerminal("{", "{")
	g.NewTerminal("}", "}")
	g.NewTerminal("if", "if")
	g.NewTerminal("print", "print")
	g.NewTerminal("bool", `true|false`).Action(func(t *c2.Token) (interface{}, error) {
		return t.Lexeme() == "true", nil
	})
	g.NewTerminal("integer", `[0-9]+`).Action(func(t *c2.Token) (interface{}, error) {
		i, err := strconv.ParseInt(t.Lexeme(), 10, 32)
		if err != nil {
			return nil, err
		}
		return int(i), nil
	})
	g.NewTerminal("float", `[0-9]+\.[0-9]+`).Action(func(t *c2.Token) (interface{}, error) {
		f, err := strconv.ParseFloat(t.Lexeme(), 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	})
	g.NewTerminal("string", `".*?"`).Action(func(t *c2.Token) (interface{}, error) {
		str := t.Lexeme()
		return str[1 : len(str)-1], nil
	})
	g.NewTerminal("+", `\+`)
	g.NewTerminal("-", `\-`)
	g.NewTerminal("*", `\*`)
	g.NewTerminal("/", `\/`)
	g.NewTerminal("varName", `\w[\w\d]*`).Action(func(t *c2.Token) (interface{}, error) {
		return t.Lexeme(), nil
	})
	g.NewTerminal("lineComment", `//.*`).Ignore()
	g.NewTerminal("blockComment", `/\*(.|\n)*?\*/`).Ignore()

	g.NewNonterminal("START").
		RHS("STATEMENTS")
	g.NewNonterminal("STATEMENTS").
		RHS("STATEMENTS STATEMENT").
		RHS("STATEMENT")
	g.NewNonterminal("STATEMENT").
		RHS("ASSIGN").
		RHS("print EXPR", func(s *c2.ASTNode) error {
			fmt.Println(s.Down(1).Data.(Value).value)
			return nil
		}).
		RHS("if EXPR {", func(s *c2.ASTNode) error {
			fmt.Println("this runs first")
			return nil
		}, "STATEMENTS }", func(s *c2.ASTNode) error {
			fmt.Println("this runs second")
			return nil
		})
	g.NewNonterminal("ASSIGN").
		RHS("STORAGE = EXPR", func(s *c2.ASTNode) error {
			storage := s.Down(0).Data.(*Value)
			*storage = s.Down(2).Data.(Value)
			return nil
		})
	g.NewNonterminal("STORAGE").
		RHS("varName", func(s *c2.ASTNode) error {
			lexeme := s.Down(0).Data.(string)
			if _, ok := variables[lexeme]; !ok {
				variables[lexeme] = &Value{}
			}
			s.Data = variables[lexeme]
			return nil
		})
	g.NewNonterminal("EXPR").
		RHS("TERM", func(s *c2.ASTNode) error {
			s.Data = s.Down(0).Data
			return nil
		}).
		RHS("EXPR ADD/SUB TERM", func(s *c2.ASTNode) error {
			lhs := s.Down(0).Data.(Value)
			rhs := s.Down(2).Data.(Value)
			op := s.Down(1).Data.(int)
			f, ok := opHandlers[opHandlerRef{op, lhs.vType, rhs.vType}]
			if !ok {
				return s.NewError(fmt.Sprintf("for binary operator %s, lhs:%s and rhs:%s are invalid types",
					opToString[op], typeToString[lhs.vType], typeToString[rhs.vType]))
			}
			f(s, lhs.value, rhs.value)
			return nil
		})
	g.NewNonterminal("TERM").
		RHS("FACTOR", func(s *c2.ASTNode) error {
			s.Data = s.Down(0).Data
			return nil
		}).
		RHS("TERM MUL/DIV FACTOR", func(s *c2.ASTNode) error {
			lhs := s.Down(0).Data.(Value)
			rhs := s.Down(2).Data.(Value)
			op := s.Down(1).Data.(int)
			f, ok := opHandlers[opHandlerRef{op, lhs.vType, rhs.vType}]
			if !ok {
				return s.NewError(fmt.Sprintf("for binary operator %s, lhs:%s and rhs:%s are invalid types",
					opToString[op], typeToString[lhs.vType], typeToString[rhs.vType]))
			}
			f(s, lhs.value, rhs.value)
			return nil
		})
	g.NewNonterminal("FACTOR").
		RHS("VALUE", func(s *c2.ASTNode) error {
			s.Data = s.Down(0).Data
			return nil
		}).
		RHS("( EXPR )", func(s *c2.ASTNode) error {
			s.Data = s.Down(1).Data
			return nil
		}).
		RHS("UNARY_OP FACTOR", func(s *c2.ASTNode) error {
			return s.NewError("unary ops not implemented")
		})
	g.NewNonterminal("VALUE").
		RHS("varName", func(s *c2.ASTNode) error {
			lexeme := s.Down(0).Data.(string)
			v, ok := variables[lexeme]
			if !ok {
				return s.NewError(fmt.Sprintf("variable (%s) is undefined", lexeme))
			}
			s.Data = *v
			return nil
		}).
		RHS("bool", func(s *c2.ASTNode) error {
			s.Data = Value{
				value: s.Down(0).Data,
				vType: vTypeBool,
			}
			return nil
		}).
		RHS("integer", func(s *c2.ASTNode) error {
			s.Data = Value{
				value: s.Down(0).Data,
				vType: vTypeInteger,
			}
			return nil
		}).
		RHS("string", func(s *c2.ASTNode) error {
			s.Data = Value{
				value: s.Down(0).Data,
				vType: vTypeInteger,
			}
			return nil
		}).
		RHS("float", func(s *c2.ASTNode) error {
			s.Data = Value{
				value: s.Down(0).Data,
				vType: vTypeInteger,
			}
			return nil
		})
	g.NewNonterminal("UNARY_OP").
		RHS("!", func(s *c2.ASTNode) error {
			return nil
		}).
		RHS("-", func(s *c2.ASTNode) error {
			return nil
		})
	g.NewNonterminal("ADD/SUB").
		RHS("+", func(s *c2.ASTNode) error {
			s.Data = opAdd
			return nil
		}).
		RHS("-", func(s *c2.ASTNode) error {
			s.Data = opSub
			return nil
		})
	g.NewNonterminal("MUL/DIV").
		RHS("*", func(s *c2.ASTNode) error {
			s.Data = opMul
			return nil
		}).
		RHS("/", func(s *c2.ASTNode) error {
			s.Data = opDiv
			return nil
		})
	return g.Build()
}
