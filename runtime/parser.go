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
	for _, t := range Tokenizer(raw) {
		fmt.Println(t)
	}

	return []Case{}
}

func Tokenizer(raw string) []Token {
	var tokens []Token
	letters := []rune(raw)

	for pos := 0; pos < len(letters); pos++ {
		switch letters[pos] {
		case rune(','):
			tokens = append(tokens, Token{kind: commaToken, lexeme: ","})
			continue

		case rune('('):
			tokens = append(tokens, Token{kind: openParToken, lexeme: "("})
			continue

		case rune(')'):
			tokens = append(tokens, Token{kind: closeParToken, lexeme: ")"})
			continue

		case rune(':'):
			if next(pos, letters).lexeme == "=" {
				tokens = append(tokens, Token{kind: defEqToken, lexeme: ":="})
				pos += 1
			} else {
				w := word(pos, letters)
				pos += len(w) - 1
				tokens = append(tokens, Token{kind: identifierToken, lexeme: w})
			}
			break

		case rune('='):
			if next(pos, letters).lexeme == ">" {
				tokens = append(tokens, Token{kind: blockOpenToken, lexeme: "=>"})
				pos += 1
			} else {
				w := word(pos, letters)
				pos += len(w) - 1
				tokens = append(tokens, Token{kind: identifierToken, lexeme: w})
			}
			break

		case rune(' '):
			fallthrough
		case rune('\t'):
			fallthrough
		case rune('\n'):
			fallthrough
		case rune('\r'):
			break

		default:
			w := word(pos, letters)
			pos += len(w) - 1
			tokens = append(tokens, Token{kind: identifierToken, lexeme: w})
			break
		}
	}

	return tokens
}

func next(pos int, letters []rune) Token {
	if pos+1 > len(letters) {
		return Token{kind: eofToken}
	} else {
		return Token{
			kind:   identifierToken,
			lexeme: string(letters[pos+1]),
		}
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
