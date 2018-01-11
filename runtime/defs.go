package runtime

import (
	"fmt"
	"strings"
)

type tokenKind string
type exprKind string
type declKind string

type Case struct {
	expr Expr
	dcls []Declaration
}

type Declaration struct {
	kind  declKind
	key   Token
	value Expr
}

type Expr struct {
	kind  exprKind
	value Token
	args  []Token
}

type Token struct {
	kind   tokenKind
	lexeme string
}

const (
	caseToken       tokenKind = "casetok"   // "case"
	defPathToken    tokenKind = "defptok"   // "path"
	blockOpenToken  tokenKind = "blockotok" // "=>"
	defEqToken      tokenKind = "defeqtok"  // ":="
	openParToken    tokenKind = "opartok"   // "("
	closeParToken   tokenKind = "cpartok"   // ")"
	commaToken      tokenKind = "commatok"  // ","
	identifierToken tokenKind = "idtok"     // [^\s]+
	eofToken        tokenKind = "eoftok"    // EOF

	call exprKind = "call"
	expr exprKind = "expr"

	path declKind = "path"
)

func (c Case) String() string {
	var dcls []string

	for _, decl := range c.dcls {
		dcls = append(dcls, fmt.Sprintf("  %s\n", decl))
	}

	return fmt.Sprintf("case %s =>\n%s", c.expr, strings.Join(dcls, ""))
}

func (d Declaration) String() string {
	switch d.kind {
	case path:
		return fmt.Sprintf("path %s := %s", d.key.lexeme, d.value)

	default:
		return "<Invalid Declaration>"
	}
}

func (e Expr) String() string {
	switch e.kind {
	case call:
		var args []string

		for _, arg := range e.args {
			args = append(args, fmt.Sprintf("%s", arg.lexeme))
		}

		return fmt.Sprintf("%s(%s)", e.value.lexeme, strings.Join(args, ", "))

	case expr:
		return fmt.Sprintf("%s", e.value)

	default:
		return "<Invalid Expression>"
	}
}
