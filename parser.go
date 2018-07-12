package c2

import (
	"io/ioutil"
	"strings"
	"regexp"
	"fmt"
	"c2/lr0"
	"bytes"
)

/*
 * Exported structs/functions
 */

// Parses a file according to symbols.
func ParseFile(filepath string) error {
	var err error
	pContent, err = ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	pInputFile = filepath
	return pParse()
}

// Parses a string according to symbols.
func ParseString(text string) error {
	r := strings.NewReader(text)
	var err error
	pContent, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	pInputFile = "text.c2l"
	return pParse()
}

type Token struct {
	id          int
	lexeme      string
	bytes       []byte
	row, column int
}

type Symbol struct {
	id    int
	token *Token
	data  interface{}
	sub   []*Symbol
}

func (s Symbol) Name() string {
	return pSymbolName(s.id)
}

/*
 * Vars for use by production actions.
 */

var onNextToken func(t *Token)
var onNextParse func(sym *Symbol, nextAction int)
var onInputAccepted func(s *Symbol)
// var onError func(e *error)

/*
 * "private" vars and funcs.
 */

var pInputFile string
var pInitialized bool
var pNewline= regexp.MustCompile("\\r?\\n")
var pContent []byte
var pIndex int
var pSize int
var pRow, pColumn= 1, 1
var pCurrentToken *Token

/*
 * Functions for use by production actions.
 */

func errorAt(r, c int, reason string) {
	errorOut(fmt.Sprintf("%v:%v | %v", r, c, reason))
}

func errorOut(reason string) {
	panic(fmt.Errorf("[Error] %v", reason))
}

// Prints out the current token to standard out.
func printCurrentToken() {
	fmt.Printf("TOKEN: %s\tVALUE: %s\n", pSymbolName(pCurrentToken.id), pCurrentToken.lexeme)
}

// Saves all tokens found into 'takenText[saveAsID]' until a token of 'untilID' is found.
func takeUntil(untilID int) string {
	revertToken := pCurrentToken
	var buffer bytes.Buffer
	nextToken := pPeekNext()
	for nextToken.id != untilID && nextToken.id != pEOF {
		pNext()
		buffer.WriteString(pCurrentToken.lexeme)
		nextToken = pPeekNext()
	}
	r, c := pRow, pColumn
	pNext()
	pCurrentToken = revertToken
	pRow, pColumn = r, c
	pIndex -= len(pCurrentToken.bytes)
	return buffer.String()
}

/*
 * "Private" parser functionality.
 */

func pProductionLength(symId, prodId int) int {
	return pRuleLengths[symId-pAugStart][prodId]
}

func pRunProductionAction(symId, prodId int, s *Symbol) {
	pRuleActions[symId-pAugStart][prodId](s)
}

func pSymbolName(sym int) string {
	return pSymToString[sym]
}

func pNext() {
	// Adjust the current pRow and pColumn.
	if pCurrentToken != nil {
		var i int
		match := pNewline.FindAllIndex(pCurrentToken.bytes, -1)
		if match != nil {
			pRow += len(match)
			i = match[len(match)-1][1]
			pColumn = 1
		}
		for _, char := range pCurrentToken.lexeme[i:] {
			if char == '\t' {
				pColumn += 4 - (pColumn-1)%4
			} else {
				pColumn++
			}
		}
		pIndex += len(pCurrentToken.bytes)
	}

	// Check for end of input.
	if pIndex >= pSize {
		pCurrentToken = &Token{
			id:     pEOF,
			bytes:  nil,
			lexeme: "",
			row:    pRow,
			column: pColumn,
		}
		return
	}

	// Try all pTerminals to find the correct token.
	for i := range pTerminals {
		match := pTerminals[i].Find(pContent[pIndex:])
		if len(match) == 0 {
			continue
		}

		pCurrentToken = &Token{
			id:     i,
			bytes:  match,
			lexeme: string(match[:]),
			row:    pRow,
			column: pColumn,
		}
		break // Final expr is 'undefined' - a match-all, so we're guaranteed to find a token.
	}
}

func pPeekNext() *Token {
	peekIndex := pIndex + len(pCurrentToken.bytes)
	if peekIndex >= pSize {
		return &Token{
			id:     pEOF,
			bytes:  nil,
			lexeme: "",
			row:    pRow,
			column: pColumn,
		}
	}

	for i := range pTerminals {
		match := pTerminals[i].Find(pContent[peekIndex:])
		if len(match) == 0 {
			continue
		}

		return &Token{
			id:     i,
			bytes:  match,
			lexeme: string(match[:]),
			row:    pRow,
			column: pColumn,
		}
	}

	return nil
}

func pParse() error {
	if !pInitialized {
		pUserInitialization()
		pInitialized = true
	}

	// Stack initialized with State = 0
	var stateStack = make([]int, 1, 10)
	var symbolStack = make([]*Symbol, 0, 10)
	var state int
	var symbol, onHoldSymbol *Symbol

	pSize = len(pContent)
	for {
		pNext()
		if onNextToken != nil {
			onNextToken(pCurrentToken)
		}
		if pIgnoreTerminals[pCurrentToken.id] {
			continue
		}
		symbol = &Symbol{id: pCurrentToken.id, token: pCurrentToken}
		if symbol.id == pUndefined {
			t := symbol.token
			errorAt(t.row, t.column, fmt.Sprintf("Unrecognized symbol: %v", t.lexeme))
		}

		action := pTerminalActions[pCurrentToken.id]
		if action != nil {
			action(symbol)
		}

		// Parse token with respect to pRuleLengths.
		ParseToken:
		e := pParseTable[state][symbol.id]

		if onNextParse != nil {
			onNextParse(symbol, e.Op)
		}

		switch e.Op {

		case lr0.OpAccept:
			fmt.Println("Input has been accepted.")
			if onInputAccepted != nil {
				onInputAccepted(symbolStack[0])
			}
			return nil

		case lr0.OpShift:
			state = e.Data
			symbolStack = append(symbolStack, symbol)
			stateStack = append(stateStack, state)
			break

		case lr0.OpReduce:
			// Remove the current state and the last input from the stack.
			rhsCount := pProductionLength(e.Data, e.ReduceRHS)
			reducedSymbols := make([]*Symbol, 0, 4)
			for _, s := range symbolStack[len(symbolStack)-rhsCount:] {
				reducedSymbols = append(reducedSymbols, s)
			}

			symbolStack = symbolStack[:len(symbolStack)-rhsCount]
			stateStack = stateStack[:len(stateStack)-rhsCount]

			state = stateStack[len(stateStack) - 1]
			onHoldSymbol = symbol
			symbol = &Symbol{id: e.Data, sub: reducedSymbols} // 'symbol' is now a non-terminal.

			pRunProductionAction(e.Data, e.ReduceRHS, symbol)
			goto ParseToken

		case lr0.OpGoto:
			state = e.Data
			symbolStack = append(symbolStack, symbol)
			symbol = onHoldSymbol
			stateStack = append(stateStack, state)
			goto ParseToken

		case lr0.OpErr:
			var s string
			if pCurrentToken.id == pEOF {
				s = "Reached EOF prematurely."
			} else {
				if symbol.token != nil {
					s += fmt.Sprintf("Found %v --> %v,", symbol.Name(), symbol.token.lexeme)
				} else {
					s += fmt.Sprintf("Found %v,", symbol.Name())
				}
			}
			s += fmt.Sprintf(" Expected one of: ")
			for sym, entry := range pParseTable[state] {
				if sym >= pTerminalCount {
					break
				}
				if entry.Op != lr0.OpErr {
					s += fmt.Sprintf("%v, ", pSymbolName(sym))
				}
			}
			errorAt(pRow, pColumn, s[:len(s) - 2])
		}
	}
	return nil
}