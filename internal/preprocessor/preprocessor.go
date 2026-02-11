package preprocessor

import (
	"embed"
	"errors"
	"path/filepath"
	"strings"
)

func parseIncludePath(path string) (string, error) {
	file := ""
	endChar := '<'
	if path[0] != '<' {
		if path[0] != '"' {
			return "", errors.New("invalid include path")
		}
		endChar = '"'
	}
	i := 1
	c := path[i]
	for c != byte(endChar) && i < len(path)-1 {
		file += string(c)
		i++
		c = path[i]
	}

	return file, nil
}

type Preprocessor struct {
	Definitions       map[string]string
	includeStatements bool
	inIfStatement     bool
	templateFS        embed.FS
}

func New(templateFS embed.FS) *Preprocessor {
	defs := make(map[string]string)

	defs["__wasm__"] = "1"
	defs["__wasm32__"] = "1"
	defs["__EMSCRIPTEN__"] = "1"
	defs["__ILP32__"] = "1"
	defs["__BIGGEST_ALIGNMENT__"] = "16"
	defs["__webc__"] = "1"

	return &Preprocessor{
		Definitions:       defs,
		includeStatements: true,
		inIfStatement:     false,
		templateFS:        templateFS,
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
				var replacedLine string = line
				for name, value := range p.Definitions {
					replacedLine = strings.ReplaceAll(replacedLine, name, value)
				}
				finalSource += replacedLine + "\n"
			}
		}

		if strings.HasPrefix(line, "#define") {
			split := strings.Split(line, " ")
			if len(split) < 2 {
				return "", errors.New("invalid preprocessor statement")
			}

			name := split[1]
			var value string
			if len(split) == 2 {
				value = "1"
			} else {
				value = strings.Join(split[2:], " ")
			}

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
		} else if strings.HasPrefix(line, "#include") {
			split := strings.Split(line, " ")
			if len(split) < 2 {
				return "", errors.New("invalid preprocessor statement")
			}

			path, err := parseIncludePath(split[1])
			if err != nil {
				return "", err
			}

			file, err := p.templateFS.ReadFile(filepath.Join("templates", "stdlib", path))
			if err != nil {
				return "", err
			}
			preProcessedFile, err := p.Parse(string(file))
			if err != nil {
				return "", err
			}

			for _, line := range strings.Split(preProcessedFile, "\n") {
				finalSource += line + "\n"
			}

		} else if strings.HasPrefix(line, "#endif") {
			if !p.inIfStatement {
				return "", errors.New("invalid preprocessor if statement")
			}
			p.includeStatements = true
		} else {

		}
	}

	return finalSource, nil
}
