package runtime

import "fmt"

type parser struct {
	pos    int
	tokens []Token
}

var eof = tok(eofToken, "<eof>")

/**
 * Servfile configuration parser
 *
 * Grammar:
 *
 *     MAIN            = match* EOF ;
 *
 *     match           = "case" expression "=>" declaration* ;
 *
 *     declaration     = "path" IDENTIFIER expression ;
 *
 *     expression      = IDENTIFIER
 *                     | IDENTIFIER "(" [IDENTIFIER ["," IDENTIFIER]*] ")" ;
 *
 *     IDENTIFIER      = [^\s]+
 *
 *
 * Sample raw input:
 *
 *     case Host(_, _, _) =>
 *       path /        git(https://github.com/minond/minond.github.io.git)
 *       path /servies git(https://github.com/minond/servies.git)
 *       path /static  dir(.)
 *       path /github  redirect(https://github.com/minond)
 *       path /ps      cmd(ps, aux)
 *       path /imdb    proxy(http://www.imdb.com:80)
 *       path /unibrow proxy(http://localhost:3001)
 *
 *
 * Sample ast output:
 *
 *     var ast = []Match{
 *       Match{
 *         expr: Expr{
 *           kind:  call,
 *           value: Token{kind: identifierToken, lexeme: "Host"},
 *           args: []Token{
 *             Token{kind: identifierToken, lexeme: "_"},
 *             Token{kind: identifierToken, lexeme: "_"},
 *             Token{kind: identifierToken, lexeme: "_"},
 *           },
 *         },
 *         dcls: []Declaration{
 *           Declaration{
 *             kind: path,
 *             key:  Token{kind: identifierToken, lexeme: "/"},
 *             value: Expr{
 *               kind:  call,
 *               value: Token{kind: identifierToken, lexeme: "git"},
 *               args: []Token{
 *                 Token{kind: identifierToken, lexeme: "https://github.com/minond/minond.github.io.git"},
 *               },
 *             },
 *           },
 *           Declaration{
 *             kind: path,
 *             key:  Token{kind: identifierToken, lexeme: "/servies"},
 *             value: Expr{
 *               kind:  call,
 *               value: Token{kind: identifierToken, lexeme: "git"},
 *               args: []Token{
 *                 Token{kind: identifierToken, lexeme: "https://github.com/minond/servies.git"},
 *               },
 *             },
 *           },
 *           Declaration{
 *             kind: path,
 *             key:  Token{kind: identifierToken, lexeme: "/static"},
 *             value: Expr{
 *               kind:  call,
 *               value: Token{kind: identifierToken, lexeme: "dir"},
 *               args: []Token{
 *                 Token{kind: identifierToken, lexeme: "."},
 *               },
 *             },
 *           },
 *           Declaration{
 *             kind: path,
 *             key:  Token{kind: identifierToken, lexeme: "/ps"},
 *             value: Expr{
 *               kind:  call,
 *               value: Token{kind: identifierToken, lexeme: "cmd"},
 *               args: []Token{
 *                 Token{kind: identifierToken, lexeme: "ps"},
 *                 Token{kind: identifierToken, lexeme: "aux"},
 *               },
 *             },
 *           },
 *         },
 *       },
 *     }
 *
 */
func Parse(raw string) []Match {
	p := parser{
		pos:    0,
		tokens: tokenize(raw),
	}

	var matches []Match

	for !p.done() {
		matches = append(matches, p.match())
	}

	return matches
}

func (p *parser) match() Match {
	if p.eat().lexeme != "case" {
		panic(fmt.Sprintf("Expecting `case` but found %s instead.",
			p.prev().lexeme))
	}

	mat := Match{}
	mat.expr = p.expression()

	if !p.matches(blockOpenToken) {
		panic(fmt.Sprintf("Expecting `=>` but found %s instead.",
			p.peek().lexeme))
	}

	for !p.done() {
		mat.dcls = append(mat.dcls, p.declaration())

		if p.peek().lexeme == "case" {
			break
		}
	}

	return mat
}

func (p *parser) declaration() Declaration {
	decl := Declaration{}

	switch p.eat().lexeme {
	case "path":
		decl.kind = path

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

	decl.value = p.expression()
	return decl
}

func (p *parser) expression() Expr {
	expr := Expr{
		kind:  expr,
		value: p.eat(),
	}

	if p.matches(openParToken) {
		var args []Token
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

func (p *parser) eat() Token {
	if p.done() {
		return eof
	} else {
		tok := p.peek()
		p.pos += 1
		return tok
	}
}

func (p parser) prev() Token {
	return p.tokens[p.pos-1]
}

func (p parser) peek() Token {
	if p.done() {
		return eof
	} else {
		return p.tokens[p.pos]
	}
}

func (p parser) done() bool {
	return p.pos >= len(p.tokens) ||
		p.tokens[p.pos].kind == eofToken
}

func tokenize(raw string) []Token {
	var tokens []Token
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

		case rune('('):
			tokens = append(tokens, tok(openParToken, "("))

		case rune(')'):
			tokens = append(tokens, tok(closeParToken, ")"))

		case rune(':'):
			if next(pos, letters).lexeme == "=" {
				tokens = append(tokens, tok(defEqToken, ":="))
				pos += 1
			} else {
				identifier()
			}

		case rune('='):
			if next(pos, letters).lexeme == ">" {
				tokens = append(tokens, tok(blockOpenToken, "=>"))
				pos += 1
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

func tok(kind tokenKind, lexeme string) Token {
	return Token{kind, lexeme}
}

func next(pos int, letters []rune) Token {
	if pos+1 > len(letters) {
		return eof
	} else {
		return tok(identifierToken, string(letters[pos+1]))
	}
}

func word(pos int, letters []rune) string {
	buff := ""

	for ; pos < len(letters); pos++ {
		switch letters[pos] {
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
