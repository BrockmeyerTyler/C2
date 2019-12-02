package c2gram

import (
	"fmt"
	"github.com/tjbrockmeyer/c2"
	"regexp"
)

type TokenAction func(t *c2.Token) (interface{}, error)

type TerminalDefinition interface {
	Ignore() TerminalDefinition
	Action(TokenAction) TerminalDefinition
}

type terminal struct {
	name   string
	regex  *regexp.Regexp
	ignore bool
	action TokenAction
}

func (g *Grammar) NewTerminal(name, regex string) TerminalDefinition {
	t := &terminal{
		name:  name,
		regex: regexp.MustCompile(fmt.Sprintf(`^(%s)`, regex)),
	}
	g.terminals = append(g.terminals, t)
	g.terminalsByName[name] = t
	return t
}

func (t *terminal) Ignore() TerminalDefinition {
	t.ignore = true
	return t
}

func (t *terminal) Action(action TokenAction) TerminalDefinition {
	t.action = action
	return t
}

func (t terminal) Find(b []byte) []byte {
	return t.regex.Find(b)
}

func (t terminal) IsIgnored() bool {
	return t.ignore
}

func (t terminal) RunAction(token *c2.Token) (interface{}, error) {
	if t.action == nil {
		return nil, nil
	}
	return t.action(token)
}
