package runtime

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
 *                     | IDENTIFIER "(" [IDENTIFIER]* ")" ;
 *
 *
 * Sample:
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
 */

var ast = []Case{
	Case{
		expr: Expr{
			kind:  call,
			value: Token{kind: caseToken, lexeme: "Host"},
			args: []Token{
				Token{kind: identifierToken, lexeme: "_"},
				Token{kind: identifierToken, lexeme: "_"},
				Token{kind: identifierToken, lexeme: "_"},
			},
		},
		dcls: []Declaration{
			Declaration{
				kind: path,
				key:  Token{kind: identifierToken, lexeme: "/"},
				value: Expr{
					kind:  call,
					value: Token{kind: identifierToken, lexeme: "git"},
					args: []Token{
						Token{kind: identifierToken, lexeme: "https://github.com/minond/minond.github.io.git"},
					},
				},
			},
			Declaration{
				kind: path,
				key:  Token{kind: identifierToken, lexeme: "/servies"},
				value: Expr{
					kind:  call,
					value: Token{kind: identifierToken, lexeme: "git"},
					args: []Token{
						Token{kind: identifierToken, lexeme: "https://github.com/minond/servies.git"},
					},
				},
			},
			Declaration{
				kind: path,
				key:  Token{kind: identifierToken, lexeme: "/static"},
				value: Expr{
					kind:  call,
					value: Token{kind: identifierToken, lexeme: "dir"},
					args: []Token{
						Token{kind: identifierToken, lexeme: "."},
					},
				},
			},
			Declaration{
				kind: path,
				key:  Token{kind: identifierToken, lexeme: "/ps"},
				value: Expr{
					kind:  call,
					value: Token{kind: identifierToken, lexeme: "cmd"},
					args: []Token{
						Token{kind: identifierToken, lexeme: "ps"},
						Token{kind: identifierToken, lexeme: "aux"},
					},
				},
			},
		},
	},
}

func Parse(raw string) []Case {
	return ast
}
