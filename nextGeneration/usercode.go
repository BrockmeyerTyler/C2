package c2

import (
	"bytes"
	"fmt"
	"regexp"
	"c2/lr0"
	"os"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// TODO: Increase from LR0.
// TODO: Advanced error handling.

/*
 * Define globally:
 */

var readSymbols = make([]*ReadSymbol, 0, 50)
var readSymbolsMap = make(map[string]*ReadSymbol)
var terminals = make([]Terminal, 0, 25)
var nonterminals = make([]Nonterminal, 0, 25)
var userInit string

type ReadSymbol struct {
	id int
	name string
	dispName string
}

func (s *ReadSymbol) IsTerminal() bool {
	return s.id < len(terminals)
}

type Terminal struct {
	name, expr, action string
	ignore bool
}

type Nonterminal struct {
	name string
	productions []Production
}

type Production struct {
	vars   []*Token
	action string
}

func WriteWithReplacedSymbolIDs(str string, buf *bytes.Buffer) {
	var insertSymID= regexp.MustCompile(`\$(\w*)\$`)
	tokens := insertSymID.FindAllStringSubmatch(str, -1)
	var i= 0

	repl := func() string {
		var ret = "$1"
		if i < len(tokens) {
			ret = fmt.Sprint(readSymbolsMap[tokens[i][1]].id)
			i++
		}
		return ret
	}
	buf.WriteString(insertSymID.ReplaceAllString(str, repl()))
}

func WriteParserGoSource() {
	assertNil := func(e error) {
		if e != nil {
			panic(e)
		}
	}

	var path string
	if !filepath.IsAbs(pInputFile) {
		wd, err := os.Getwd()
		assertNil(err)
		path = filepath.Join(wd, pInputFile)
	} else {
		path = pInputFile
	}
	filename := filepath.Base(path)

	var noExt string
	dotIndex := strings.LastIndex(filename, ".")
	if dotIndex != -1 {
		noExt = filename[:dotIndex]
	} else {
		noExt = filename
	}

	dir := filepath.Dir(path)
	langSrc := filepath.Join(dir, noExt + ".go")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}

	f, err := os.Create(langSrc)
	assertNil(err)
	_, err = f.WriteString(fmt.Sprintf("package %v\n\nimport(\n\t\"fmt\"\n\t\"regexp\"\n)\n\n", noExt))
	assertNil(err)
	_, err = f.Write(generateInitialization())
	assertNil(err)
	_, err = f.Write(generateSymbols())
	assertNil(err)
	_, err = f.Write(generateParseTable())
	assertNil(err)
	f.Close()
}

func GenerateFiles(dir, language string) {
	/*if !filepath.IsAbs(dir) {
		cwd, _ := os.Getwd()
		dir = filepath.Join(cwd, dir)
	}*/
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	mainDir := filepath.Join(dir, "main")
	if _, err := os.Stat(mainDir); os.IsNotExist(err) {
		os.Mkdir(mainDir, os.ModePerm)
	}
	langFile := language
	if hasMatch, err := regexp.MatchString(`\.[cC]2[lL]$`, langFile); err == nil && !hasMatch {
		langFile += ".c2l"
	}

	ioutil.WriteFile(filepath.Join(dir, langFile),
		pEmptyC2L,
		os.ModePerm)
	ioutil.WriteFile(filepath.Join(dir, "parser.go"),
		[]byte("package " + language + pParserFile),
		os.ModePerm)
	ioutil.WriteFile(filepath.Join(mainDir, "main.go"),
		[]byte(fmt.Sprintf(pMainFile, language, language)),
		os.ModePerm)
}

func generateInitialization() []byte {
	var buf bytes.Buffer
	buf.WriteString("// Initialization:\n\nfunc pUserInitialization() {\n")
	WriteWithReplacedSymbolIDs(userInit, &buf)
	buf.WriteString("\n}\n\n")
	return buf.Bytes()
}

func generateSymbols() []byte {
	var newline= regexp.MustCompile(`\r?\n`)
	var buf bytes.Buffer

	// Write symbol constants.
	buf.WriteString(fmt.Sprintf("const (\n"))
	undef := len(terminals)
	ruleStart := undef + 4
	buf.WriteString(fmt.Sprintf(
		"\tpUndefined = %v"+
			"\n\tpLambda = %v"+
			"\n\tpEOF = %v"+
			"\n\tpTerminalCount = %v"+
			"\n\tpAugStart = %v"+
			"\n\tpSymbolCount = %v\n)",
		undef, undef+1, undef+2, undef+3, undef+3, ruleStart+len(nonterminals)))

	// Write symbol-to-string slice.
	buf.WriteString("\n\nvar pSymToString = []string {\n")
	for i := range terminals {
		buf.WriteString(fmt.Sprintf("\t\"%v\",\n", readSymbols[i].dispName))
	}
	buf.WriteString(`
	"undefined",
	"~",
	"$",

	"AUG_START",
`)
	for i := range nonterminals {
		buf.WriteString(fmt.Sprintf("\t\"%v\",\n", nonterminals[i].name))
	}
	buf.WriteString("}")

	// Terminal regex
	buf.WriteString("\n\nvar pTerminals = []*regexp.Regexp {\n")
	for i := range terminals {
		buf.WriteString(fmt.Sprintf("\tregexp.MustCompile(`^(?:%v)`),\n", terminals[i].expr))
	}
	buf.WriteString("\tregexp.MustCompile(`^.`),\n") // Create undefined as the last regex.
	buf.WriteString("}")

	// Terminal actions
	buf.WriteString("\n\nvar pTerminalActions = map[int]func(s *Symbol) {\n")
	for i := range terminals {
		if terminals[i].action != "" {
			buf.WriteString(fmt.Sprintf("\t%v: func(s *Symbol) { ", i))
			WriteWithReplacedSymbolIDs(terminals[i].action, &buf)
			buf.WriteString(" },\n")
		}
	}
	buf.WriteString("}")

	// Terminal ignores
	buf.WriteString("\n\nvar pIgnoreTerminals = map[int]bool {\n")
	for i := range terminals {
		if terminals[i].ignore {
			buf.WriteString(fmt.Sprintf("\t%v: true,\n", i))
		}
	}
	buf.WriteString("}")

	// Rules
	buf.WriteString("\n\nvar pRuleLengths = [][]int{\n")
	buf.WriteString(fmt.Sprintf("\t{ 2 },\n")) // AugStart
	for i := range nonterminals {
		buf.WriteString("\t{ ")
		for j := len(nonterminals[i].productions) - 1; j >= 0; j-- {
			buf.WriteString(fmt.Sprintf("%v, ", len(nonterminals[i].productions[j].vars)))
		}
		buf.WriteString("},\n")
	}
	buf.WriteString("}")

	// Rule actions
	buf.WriteString("\n\nvar pRuleActions = [][]func(s *Symbol) {\n")
	buf.WriteString("\t{ func(s *Symbol) {}, },\n") // AugStart
	for i := range nonterminals {
		buf.WriteString("\t{\n")
		for j := len(nonterminals[i].productions) - 1; j >= 0; j-- {
			p := nonterminals[i].productions[j]

			buf.WriteString("\t\tfunc(s *Symbol) {")
			if p.action != "" {
				if newline.MatchString(p.action) {
					buf.WriteString("\n\t\t")
					WriteWithReplacedSymbolIDs(p.action, &buf)
					buf.WriteString("\n\t\t")
				} else {
					buf.WriteString(" ")
					WriteWithReplacedSymbolIDs(p.action, &buf)
					buf.WriteString(" ")
				}
			}
			buf.WriteString("},\n")
		}
		buf.WriteString("\t},\n")
	}
	buf.WriteString("}\n\n")

	return buf.Bytes()
}

func generateParseTable() []byte {
	augStart := len(terminals) + 3
	eof := augStart - 1
	symCount := augStart + 1 + len(nonterminals)

	productions := make([][][]int, symCount-augStart)
	productions[0] = [][]int{{augStart + 1, eof}}
	for i := range nonterminals {
		productions[i+1] = make([][]int, len(nonterminals[i].productions))
		rhsCount := len(nonterminals[i].productions)
		for j := range nonterminals[i].productions {
			productions[i+1][rhsCount-j-1] = make([]int, len(nonterminals[i].productions[j].vars))
		}
		for j := range nonterminals[i].productions {
			p := nonterminals[i].productions[j]
			symbolCount := len(p.vars)
			for k := range p.vars {
				id := readSymbolsMap[p.vars[k].lexeme].id
				if id < len(terminals) {
					productions[i+1][rhsCount-j-1][symbolCount-k-1] = id
				} else {
					productions[i+1][rhsCount-j-1][symbolCount-k-1] = id + 4
				}
			}
		}
	}
	names := make([]string, symCount)
	for i := 0; i < symCount; i++ {
		if i < len(terminals) {
			names[i] = readSymbols[i].name
		} else {
			if i == eof {
				names[i] = "$"
			} else if i == eof-1 {
				names[i] = "~"
			} else if i == eof-2 {
				names[i] = "undefined"
			} else if i == augStart {
				names[i] = "AUG_START"
			} else {
				names[i] = readSymbols[i-4].name
			}
		}
	}

	isNonterminal := func(a int) bool {
		return a >= augStart && a < symCount
	}
	getSymName := func(a int) string {
		return names[a]
	}
	getProduction := func(a, b int) []int {
		return productions[a-augStart][b]
	}
	getProductions := func(a int) [][]int {
		return productions[a-augStart]
	}

	table := lr0.GenerateParseTable(isNonterminal, getProduction, getProductions, getSymName, augStart, eof, symCount)

	var buf bytes.Buffer
	buf.WriteString("type LR0ParseTableEntry struct { Data, ReduceRHS, Op int }\nvar pParseTable = ")
	buf.Write(lr0.BytePrintCode(table))
	buf.WriteString("\n\n")
	return buf.Bytes()
}