package serv

import (
	"fmt"
	"net/http"
	"strings"
)

type tokenKind string
type exprKind string
type declKind string
type routeKind string

type Match struct {
	expr  Expr
	decls []Declaration
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

type Server struct {
	Match  func(http.Request) bool
	Routes []Route
	Mux    *http.ServeMux
}

type Route struct {
	Kind routeKind
	Path string
	Data string
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
	expr exprKind = "expr"

	path declKind = "path"
	def  declKind = "def"

	cmdRoute      routeKind = "cmd"      // Wants a command string
	dirRoute      routeKind = "dir"      // Wants a directory
	gitRoute      routeKind = "git"      // Wants a git url
	proxyRoute    routeKind = "proxy"    // Wants url:port?
	redirectRoute routeKind = "redirect" // Wants a url
)

func (route Route) IsCmd() bool {
	return route.Kind == cmdRoute
}

func (route Route) IsDir() bool {
	return route.Kind == dirRoute
}

func (route Route) IsGit() bool {
	return route.Kind == gitRoute
}

func (route Route) IsProxy() bool {
	return route.Kind == proxyRoute
}

func (route Route) IsRedirect() bool {
	return route.Kind == redirectRoute
}

func (m Match) String() string {
	var decls []string

	for _, decl := range m.decls {
		decls = append(decls, fmt.Sprintf("  %s\n", decl))
	}

	return fmt.Sprintf("case %s =>\n%s", m.expr, strings.Join(decls, ""))
}

func (d Declaration) String() string {
	switch d.kind {
	case path:
		return fmt.Sprintf("path %s %s", d.key.lexeme, d.value)

	case def:
		return fmt.Sprintf("def %s %s", d.key.lexeme, d.value)

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

	case list:
		var items []string

		for _, item := range e.args {
			items = append(items, fmt.Sprintf("%s", item.lexeme))
		}

		return fmt.Sprintf("[%s]", strings.Join(items, " "))

	case expr:
		return fmt.Sprintf("%s", e.value.lexeme)

	default:
		return "<Invalid Expression>"
	}
}

func (e Expr) Value() string {
	if e.kind == expr {
		return e.value.lexeme
	} else {
		return ""
	}
}

func (e Expr) Values() []string {
	if e.kind == list {
		var vals []string

		for _, v := range e.args {
			vals = append(vals, v.lexeme)
		}

		return vals
	} else {
		return []string{}
	}
}
