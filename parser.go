package c2

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/c2/c2gram"
	"io/ioutil"
	"regexp"
	"strings"
)

const (
	OpErr = iota
	OpShift
	OpReduce
	OpGoto
	OpAccept
)

var opNames = map[byte]string{
	OpErr:    "err",
	OpShift:  "sft",
	OpReduce: "rdc",
	OpGoto:   "gto",
	OpAccept: "acc",
}

var newlineRegex = regexp.MustCompile(`\r?\n`)

type Parser struct {
	OnNextToken func(t *Token)
	OnNextParse func(n *ASTNode, nextAction byte)
	grammar     c2gram.Assembled
	file        *File
	token       *Token
	parseTable  ParseTable
}

func NewParser(g c2gram.Assembled, t ParseTable) *Parser {
	return &Parser{
		grammar:    g,
		parseTable: t,
	}
}

// Parses a file according to symbols.
func (p *Parser) ReadFile(filepath string) error {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	p.file = &File{
		name:    filepath,
		content: content,
		size:    len(content),
	}
	return p.parse()
}

// Parses a string according to symbols.
func (p *Parser) ReadString(text string) error {
	r := strings.NewReader(text)
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	p.file = &File{
		name:    "<string>",
		content: content,
		size:    len(content),
	}
	return p.parse()
}

// Move to the next token in the input.
func (p *Parser) Next() {
	t := p.PeekNext()
	// Adjust the current pRow and pColumn.
	if p.token != nil {
		file := p.file
		var i int
		match := newlineRegex.FindAllIndex(p.token.bytes, -1)
		if match != nil {
			file.row += len(match)
			i = match[len(match)-1][1]
			file.column = 1
		}
		for _, char := range p.token.lexeme[i:] {
			if char == '\t' {
				file.column += 4 - (file.column-1)%4
			} else {
				file.column++
			}
		}
		file.index += len(p.token.bytes)
	}
	p.token = t
}

func (p *Parser) PeekNext() *Token {
	file := p.file
	var skipLen int
	if p.token != nil {
		skipLen = len(p.token.bytes)
	}
	peekIndex := file.index + skipLen
	if peekIndex >= file.size {
		return &Token{
			symbol: p.grammar.Symbols[p.grammar.EndOfFile].(c2gram.Terminal),
			bytes:  nil,
			lexeme: "",
			row:    file.row,
			column: file.column,
		}
	}

	for tId := 0; tId < p.grammar.AugmentedStart; tId++ {
		t := p.grammar.Symbols[tId].(c2gram.Terminal)
		match := t.Find(file.content[peekIndex:])
		if len(match) == 0 {
			continue
		}

		return &Token{
			symbol: p.grammar.Symbols[tId].(c2gram.Terminal),
			bytes:  match,
			lexeme: string(match[:]),
			file:   file.name,
			row:    file.row,
			column: file.column,
		}
	}
	return nil
}

func (p *Parser) parse() error {
	ast, err := p.buildAST()
	if err != nil {
		return errors.WithMessage(err, "error while parsing files and building AST")
	}
	if err = ast.Traverse(); err != nil {
		return errors.WithMessage(err, "error while traversing AST")
	}
	return err
}

func (p *Parser) buildAST() (*ASTNode, error) {
	// Stack initialized with State = 0
	var stateStack = make([]int, 1, 10)
	var symbolStack = make([]*ASTNode, 0, 10)
	var state int
	var node, onHoldNode *ASTNode

	for {
		p.Next()
		if p.OnNextToken != nil {
			p.OnNextToken(p.token)
		}
		symbol := p.token.symbol
		if symbol.IsIgnored() {
			continue
		}
		node = &ASTNode{
			symbol: symbol,
			token:  p.token,
			file:   p.token.file,
			row:    p.token.row,
			column: p.token.column,
		}
		if symbol.ID() == p.grammar.Undefined {
			return nil, node.NewError(fmt.Sprintf("unrecognized symbol: '%v'", p.token.lexeme))
		}

		if output, err := symbol.RunAction(p.token); err != nil {
			return nil, errors.WithMessage(err, "error running terminal action")
		} else {
			node.Data = output
		}

		// Parse token with respect to pRuleLengths.
	ParseToken:
		e := p.parseTable[state][node.symbol.ID()]

		if p.OnNextParse != nil {
			p.OnNextParse(node, e.Op)
		}

		switch e.Op {

		case OpAccept:
			fmt.Println("Input has been accepted.")
			return symbolStack[0], nil

		case OpShift:
			state = e.Data
			symbolStack = append(symbolStack, node)
			stateStack = append(stateStack, state)
			break

		case OpReduce:
			// Remove the current state and the last input from the stack.
			production := p.grammar.Productions[e.Data]
			rhsCount := len(production.RHS)
			reducedNodes := make([]*ASTNode, 0, 4)
			for _, s := range symbolStack[len(symbolStack)-rhsCount:] {
				reducedNodes = append(reducedNodes, s)
			}

			symbolStack = symbolStack[:len(symbolStack)-rhsCount]
			stateStack = stateStack[:len(stateStack)-rhsCount]

			state = stateStack[len(stateStack)-1]
			onHoldNode = node
			node = &ASTNode{
				symbol:     p.grammar.Symbols[production.LHS],
				down:       reducedNodes,
				production: production,
				file:       reducedNodes[0].file,
				row:        reducedNodes[0].row,
				column:     reducedNodes[0].column,
			}
			for _, n := range node.down {
				n.up = node
			}
			// if f := production.Actions[len(production.RHS)]; f != nil {
			// 	if err := f(node); err != nil {
			// 		return nil, err
			// 	}
			// }
			goto ParseToken

		case OpGoto:
			state = e.Data
			symbolStack = append(symbolStack, node)
			node = onHoldNode
			stateStack = append(stateStack, state)
			goto ParseToken

		case OpErr:
			var s string
			if symbol.ID() == p.grammar.EndOfFile {
				s += "Reached EOF prematurely."
			} else {
				if node.symbol.ID() < p.grammar.AugmentedStart {
					s += fmt.Sprintf("Found %v --> %v,", node.symbol.Name(), p.token.lexeme)
				} else {
					s += fmt.Sprintf("Found %v,", node.symbol.Name())
				}
			}
			expectedSymbols := make([]string, 0, 8)
			for sym, entry := range p.parseTable[state] {
				if sym >= p.grammar.AugmentedStart {
					break
				}
				if entry.Op != OpErr {
					expectedSymbols = append(expectedSymbols, p.grammar.Symbols[sym].Name())
				}
			}
			if len(expectedSymbols) > 0 {
				s += fmt.Sprint("expected one of: ", strings.Join(expectedSymbols, ", "))
			} else {
				s += "but there are no symbols expected to be found (check the grammar for errors)"
			}
			return nil, node.NewError(s)
		}
	}
}
