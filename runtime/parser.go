package runtime

import "fmt"

/**
 * Servfile configuration parser
 *
 * Grammar:
 *
 *     MAIN            = case* EOF ;
 *
 *     case            = "case" expression "=>" declaration* ;
 *
 *     declaration     = "path" IDENTIFIER ":=" expression ;
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
 *       path /        := git(https://github.com/minond/minond.github.io.git)
 *       path /servies := git(https://github.com/minond/servies.git)
 *       path /static  := dir(.)
 *       path /github  := redirect(https://github.com/minond)
 *       path /ps      := cmd(ps, aux)
 *       path /imdb    := proxy(http://www.imdb.com:80)
 *       path /unibrow := proxy(http://localhost:3001)
 *
 *
 * Sample ast output:
 *
 *     var ast = []Case{
 *       Case{
 *         expr: Expr{
 *           kind:  call,
 *           value: Token{kind: caseToken, lexeme: "Host"},
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
func Parse(raw string) []Case {
	tokens := tokenize(raw)

	for _, t := range tokens {
		fmt.Println(t)
	}

	return []Case{}
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
		return tok(eofToken, "")
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
