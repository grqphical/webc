package preprocessor

import (
	"errors"
	"strings"
)

type Preprocessor struct {
	Definitions map[string]string
}

func New() *Preprocessor {
	defs := make(map[string]string)

	defs["__wasm__"] = "1"
	defs["__webc__"] = "1"

	return &Preprocessor{
		Definitions: defs,
	}
}

func (p *Preprocessor) Parse(sourceCode string) error {
	for _, line := range strings.Split(sourceCode, "\n") {
		// remove trailing whitespace
		line := strings.TrimLeft(line, " \t")

		if !(line[0] == '#') {
			continue
		}

		if strings.HasPrefix(line, "#define") {
			split := strings.Split(line, " ")
			if len(split) < 3 {
				return errors.New("invalid preprocessor statement")
			}

			name := split[1]
			value := strings.Join(split[2:], " ")

			p.Definitions[name] = value
		}
	}

	return nil
}
