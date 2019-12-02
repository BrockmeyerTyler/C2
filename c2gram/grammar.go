package c2gram

type Grammar struct {
	terminals          []*terminal
	nonTerminals       []*nonTerminal
	terminalsByName    map[string]*terminal
	nonTerminalsByName map[string]*nonTerminal
}

func New() *Grammar {
	return &Grammar{
		terminals:          make([]*terminal, 0, 20),
		nonTerminals:       make([]*nonTerminal, 0, 20),
		terminalsByName:    make(map[string]*terminal, 20),
		nonTerminalsByName: make(map[string]*nonTerminal, 20),
	}
}
