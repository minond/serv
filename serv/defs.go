package serv

import (
	"fmt"
	"net/http"
	"strings"
)

type tokenKind string
type exprKind string
type declKind string

type expr struct {
	kind exprKind
	val  token
	args []token
}

type match struct {
	expr  expr
	decls []declaration
}

type declaration struct {
	kind declKind
	key  token
	val  expr
}

type token struct {
	kind   tokenKind
	lexeme string
}

type server struct {
	Match  func(http.Request) bool
	Mux    *http.ServeMux
	routes []route
}

type route struct {
	handler handlerDef
	path    string
	data    string
}

const (
	blockOpenToken  tokenKind = "blockotok" // "=>"
	defEqToken      tokenKind = "defeqtok"  // ":="
	openSqrToken    tokenKind = "osqrtok"   // "["
	closeSqrToken   tokenKind = "csqrtok"   // "]"
	openParToken    tokenKind = "opartok"   // "("
	closeParToken   tokenKind = "cpartok"   // ")"
	commaToken      tokenKind = "commatok"  // ","
	identifierToken tokenKind = "idtok"     // [^\s,()]+
	eofToken        tokenKind = "eoftok"    // EOF

	call exprKind = "call"
	list exprKind = "list"
	exp  exprKind = "exp"

	path declKind = "path"
	def  declKind = "def"
)

func (m match) String() string {
	var decls []string

	for _, decl := range m.decls {
		decls = append(decls, fmt.Sprintf("  %s\n", decl))
	}

	return fmt.Sprintf("case %s =>\n%s", m.expr, strings.Join(decls, ""))
}

func (d declaration) String() string {
	switch d.kind {
	case path:
		return fmt.Sprintf("path %s %s", d.key.lexeme, d.val)

	case def:
		return fmt.Sprintf("def %s %s", d.key.lexeme, d.val)

	default:
		return "<Invalid Declaration>"
	}
}

func (e expr) String() string {
	switch e.kind {
	case call:
		var args []string

		for _, arg := range e.args {
			args = append(args, fmt.Sprintf("%s", arg.lexeme))
		}

		return fmt.Sprintf("%s(%s)", e.val.lexeme, strings.Join(args, ", "))

	case list:
		var items []string

		for _, item := range e.args {
			items = append(items, fmt.Sprintf("%s", item.lexeme))
		}

		return fmt.Sprintf("[%s]", strings.Join(items, " "))

	case exp:
		return fmt.Sprintf("%s", e.val.lexeme)

	default:
		return "<Invalid Expression>"
	}
}

func (e expr) Value() string {
	if e.kind == exp {
		return e.val.lexeme
	}

	return ""
}

func (e expr) Values() []string {
	if e.kind == list {
		var vals []string

		for _, v := range e.args {
			vals = append(vals, v.lexeme)
		}

		return vals
	}

	return []string{}
}
