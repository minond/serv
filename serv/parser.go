package serv

import (
	"fmt"
	"strings"
)

type parser struct {
	pos    int
	tokens []token
}

var eof = tok(eofToken, "<eof>")

// Parse takes the string configuration, parses it, and returns a slice
// of declarations and matchers.
func Parse(raw string) ([]declaration, []match) {
	p := parser{
		pos:    0,
		tokens: tokenize(preprocessor(raw)),
	}

	var decls []declaration
	var matches []match

	for !p.done() {
		if p.peek().lexeme == "case" {
			matches = append(matches, p.match())
		} else {
			decls = append(decls, p.declaration())
		}
	}

	return decls, matches
}

func (p *parser) match() match {
	if p.eat().lexeme != "case" {
		panic(fmt.Sprintf("Expecting `case` but found %s instead.",
			p.prev().lexeme))
	}

	mat := match{}
	mat.expr = p.expression()

	if !p.matches(blockOpenToken) {
		panic(fmt.Sprintf("Expecting `=>` but found %s instead.",
			p.peek().lexeme))
	}

	for !p.done() {
		mat.decls = append(mat.decls, p.declaration())

		if p.peek().lexeme == "case" {
			break
		}
	}

	return mat
}

func (p *parser) declaration() declaration {
	decl := declaration{}

	switch p.eat().lexeme {
	case "path":
		decl.kind = path

	case "def":
		decl.kind = def

	default:
		panic(fmt.Sprintf("Invalid declaration type: %s",
			p.prev().lexeme))
	}

	if p.matches(identifierToken) {
		decl.key = p.prev()
	} else {
		panic(fmt.Sprintf("Expecting an identifier but found %s instead",
			p.peek().kind))
	}

	decl.val = p.expression()
	return decl
}

func (p *parser) expression() expr {
	expr := expr{kind: exp}

	// Handles "[" IDENTIFIER* "]"
	if p.matches(openSqrToken) {
		var items []token
		expr.kind = list

		for !p.matches(closeSqrToken) {
			items = append(items, p.eat())
			expr.args = items
		}
	} else {
		// Handles IDENTIFIER
		//       | IDENTIFIER "(" [IDENTIFIER ["," IDENTIFIER]*] ")" ;
		expr.val = p.eat()

		if p.matches(openParToken) {
			var args []token
			expr.kind = call

		arg:
			if p.matches(closeParToken) {
				return expr
			}

			if p.matches(identifierToken) {
				args = append(args, p.prev())
				expr.args = args
			} else {
				panic(fmt.Sprintf("Expecting an identifier but found %s",
					p.peek().kind))
			}

			if p.matches(commaToken) {
				goto arg
			} else if !p.matches(closeParToken) {
				panic(fmt.Sprintf("Expecting a closing paren but found %s",
					p.peek().kind))
			}
		}
	}

	return expr
}

func (p *parser) matches(kinds ...tokenKind) bool {
	for _, kind := range kinds {
		if p.peek().kind == kind {
			p.eat()
			return true
		}
	}

	return false
}

func (p *parser) eat() token {
	if p.done() {
		return eof
	}

	tok := p.peek()
	p.pos++
	return tok
}

func (p parser) prev() token {
	return p.tokens[p.pos-1]
}

func (p parser) peek() token {
	if p.done() {
		return eof
	}

	return p.tokens[p.pos]
}

func (p parser) done() bool {
	return p.pos >= len(p.tokens) ||
		p.tokens[p.pos].kind == eofToken
}

// In charge of prepping raw text for the tokenizer. Right now this just means
// removing comments but if we wanted to add macros, they could be handled
// here.
func preprocessor(raw string) string {
	var processed []string

	startsWith := func(prefix, str string) bool {
		if len(str) == 0 {
			return false
		}

		return strings.HasPrefix(strings.TrimSpace(str), prefix)
	}

	for _, line := range strings.Split(raw, "\n") {
		if startsWith("#", line) {
			continue
		}

		processed = append(processed, line)
	}

	return strings.Join(processed, "\n")
}

func tokenize(raw string) []token {
	var tokens []token
	letters := []rune(raw)
	pos := 0

	identifier := func() {
		w := word(pos, letters)
		pos += len(w) - 1
		tokens = append(tokens, tok(identifierToken, w))
	}

	for ; pos < len(letters); pos++ {
		switch letters[pos] {
		case rune(','):
			tokens = append(tokens, tok(commaToken, ","))

		case rune('['):
			tokens = append(tokens, tok(openSqrToken, "["))

		case rune(']'):
			tokens = append(tokens, tok(closeSqrToken, "]"))

		case rune('('):
			tokens = append(tokens, tok(openParToken, "("))

		case rune(')'):
			tokens = append(tokens, tok(closeParToken, ")"))

		case rune(':'):
			if next(pos, letters).lexeme == "=" {
				tokens = append(tokens, tok(defEqToken, ":="))
				pos++
			} else {
				identifier()
			}

		case rune('='):
			if next(pos, letters).lexeme == ">" {
				tokens = append(tokens, tok(blockOpenToken, "=>"))
				pos++
			} else {
				identifier()
			}

		case rune(' '):
		case rune('\t'):
		case rune('\n'):
		case rune('\r'):

		default:
			identifier()
		}
	}

	return tokens
}

func tok(kind tokenKind, lexeme string) token {
	return token{kind, lexeme}
}

func next(pos int, letters []rune) token {
	if pos+1 > len(letters) {
		return eof
	}

	return tok(identifierToken, string(letters[pos+1]))
}

func word(pos int, letters []rune) string {
	buff := ""

	for ; pos < len(letters); pos++ {
		switch letters[pos] {
		case rune('['):
			fallthrough
		case rune(']'):
			fallthrough
		case rune('('):
			fallthrough
		case rune(')'):
			fallthrough
		case rune(','):
			fallthrough
		case rune(' '):
			fallthrough
		case rune('\t'):
			fallthrough
		case rune('\n'):
			fallthrough
		case rune('\r'):
			return buff

		default:
			buff += string(letters[pos])
		}
	}

	return buff
}
