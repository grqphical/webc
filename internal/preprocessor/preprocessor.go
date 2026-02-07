package preprocessor

import (
	"errors"
	"strings"
)

type Preprocessor struct {
	Definitions       map[string]string
	includeStatements bool
	inIfStatement     bool
}

func New() *Preprocessor {
	defs := make(map[string]string)

	defs["__wasm__"] = "1"
	defs["__webc__"] = "1"

	return &Preprocessor{
		Definitions:       defs,
		includeStatements: true,
		inIfStatement:     false,
	}
}

func (p *Preprocessor) Parse(sourceCode string) (string, error) {
	finalSource := ""
	for _, line := range strings.Split(sourceCode, "\n") {
		// remove trailing whitespace
		line := strings.TrimLeft(line, " \t")

		if len(line) < 1 {
			continue
		}

		if !(line[0] == '#') {
			if p.includeStatements {
				finalSource += line + "\n"
			}
		}

		if strings.HasPrefix(line, "#define") {
			split := strings.Split(line, " ")
			if len(split) < 3 {
				return "", errors.New("invalid preprocessor statement")
			}

			name := split[1]
			value := strings.Join(split[2:], " ")

			p.Definitions[name] = value
		} else if strings.HasPrefix(line, "#ifdef") {
			p.inIfStatement = true
			split := strings.Split(line, " ")
			if len(split) < 2 {
				return "", errors.New("invalid preprocessor statement")
			}

			definition := split[1]
			if _, ok := p.Definitions[definition]; !ok {
				p.includeStatements = false
			}
		} else if strings.HasPrefix(line, "#ifndef") {
			p.inIfStatement = true
			split := strings.Split(line, " ")
			if len(split) < 2 {
				return "", errors.New("invalid preprocessor statement")
			}

			definition := split[1]
			if _, ok := p.Definitions[definition]; ok {
				p.includeStatements = false
			}
		} else if strings.HasPrefix(line, "#else") {
			p.includeStatements = !p.includeStatements
		} else if strings.HasPrefix(line, "#endif") {
			if !p.inIfStatement {
				return "", errors.New("invalid preprocessor if statement")
			}
			p.includeStatements = true
		}
	}

	return finalSource, nil
}
