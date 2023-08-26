package fieldfilter

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Selection map[string]Selection

// Compile the given str to selection
//
// The syntax is loosely based on XPath:
//
// a       select a field 'a'
// a,b,c   comma-separated list will select multiple fields
// a/b/c   path will select a field from its parent
// a(b,c)  sub-selection will select many fields from a parent
// a/*/c   the star * wildcard will select all items in a field
// a,b/c(d,e(f,g/h)),i
//
// nolint:gomnd
func Compile(str string) (Selection, error) {
	if !utf8.ValidString(str) {
		return nil, fmt.Errorf("invalid fields")
	}
	tokens := make([]string, 0, len(str)/2)
	var state uint8
	var three, four uint8
	three, four = 3, 4
	var b strings.Builder
	b.Grow(len(str) / 2)
	for _, r := range str {
		switch r {
		case ',':
			switch state { // nolint:gomnd
			case four: // nolint:gomnd
				tokens = append(tokens, b.String())
			case three: // nolint:gomnd
				// nothing to do
			default:
				return nil, fmt.Errorf("invalid char before ','")
			}
			tokens = append(tokens, ",")
			b.Reset()
			state = 1
		case '/':
			if state != four {
				return nil, fmt.Errorf("invalid char before '/'")
			}
			tokens = append(tokens, b.String())
			tokens = append(tokens, "/")
			b.Reset()
			state = 2
		case '(':
			if state != four {
				return nil, fmt.Errorf("invalid char before '('")
			}
			tokens = append(tokens, b.String())
			tokens = append(tokens, "(")
			b.Reset()
			state = 2
		case ')':
			switch state {
			case 4:
				tokens = append(tokens, b.String())
			case 3:
				// nothing to do
			default:
				return nil, fmt.Errorf("invalid char before ')'")
			}
			tokens = append(tokens, ")")
			b.Reset()
			state = 3
		default:
			if state == 3 {
				return nil, fmt.Errorf("invalid ')' before a char")
			}
			b.WriteRune(r)
			state = 4
		}
	}

	switch state {
	case four:
		tokens = append(tokens, b.String())
	case three:
		// nothing to do
	default:
		return nil, fmt.Errorf("invalid end")
	}

	node := make(Selection)
	err := buildSelection(tokens, node)
	return node, err
}

func buildSelection(tokens []string, root Selection) error {
	if len(tokens) == 0 {
		return nil
	}

	var child Selection
	node := root
	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case ",":
			node = root
		case "/":
			node = child
		case "(":
			end := findCloseIndex(tokens, i+1)
			if end == -1 {
				return fmt.Errorf("sub-selector not close")
			}
			if err := buildSelection(tokens[i+1:end], child); err != nil {
				return err
			}
			i = end
		case ")":
			return fmt.Errorf("invalid field char: ')'")
		default:
			child = make(Selection)
			node[tokens[i]] = child
		}
	}
	return nil
}

func findCloseIndex(tokens []string, start int) int {
	for i := len(tokens) - 1; i >= start; i-- {
		if tokens[i] == ")" {
			return i
		}
	}
	return -1
}
