package serv

import (
	"net/http"
	"strings"
)

type Environement struct {
	Matchers     map[string]MatcherDef
	Handlers     map[string]HandlerDef
	Declarations []Declaration
}

type RuntimeValue struct {
	Value string
}

type HandlerDef struct {
	Arity       int
	Constructor func(Route, *http.ServeMux)
}

type MatcherDef struct {
	Arity       int
	Constructor func(...string) Matcher
}

type Matcher interface {
	Match(http.Request) bool
}

type NullMatcher struct {
}

type HostMatcher struct {
	Subdomain RuntimeValue
	Domain    RuntimeValue
	Tld       RuntimeValue
}

func Value(val string) RuntimeValue {
	return RuntimeValue{Value: val}
}

func (v RuntimeValue) Equals(other string) bool {
	if v.Value == "_" {
		return true
	} else {
		return v.Value == other
	}
}

func (n NullMatcher) Match(r http.Request) bool {
	return false
}

func (h HostMatcher) Match(r http.Request) bool {
	parts := strings.Split(r.Host, ".")
	subdomain := ""
	domain := ""
	tld := ""

	switch len(parts) {
	case 1:
		domain = parts[0]

	case 2:
		domain = parts[0]
		tld = parts[1]

	case 3:
		subdomain = parts[0]
		domain = parts[1]
		tld = parts[2]
	}

	return h.Subdomain.Equals(subdomain) &&
		h.Domain.Equals(domain) &&
		h.Tld.Equals(tld)
}

func NewEnvironment(decls []Declaration) Environement {
	return Environement{
		Declarations: decls,
		Matchers: map[string]MatcherDef{
			"Host": {
				Arity: 3,
				Constructor: func(args ...string) Matcher {
					return HostMatcher{
						Subdomain: Value(args[0]),
						Domain:    Value(args[1]),
						Tld:       Value(args[2]),
					}
				},
			},
		},

		Handlers: map[string]HandlerDef{
			"git": {
				Arity: 1,
				Constructor: func(route Route, mux *http.ServeMux) {
					AssertGitRepo(route.Data)
					SetGitHandler(mux, route)
					go PullGitRepoInterval(route.Data)
				},
			},
			"dir": {
				Arity: 1,
				Constructor: func(route Route, mux *http.ServeMux) {
					AssertDir(route.Data)
					SetDirHandler(mux, route)
				},
			},
			"redirect": {
				Arity: 1,
				Constructor: func(route Route, mux *http.ServeMux) {
					SetRedirectHandler(mux, route)
				},
			},
			"cmd": {
				Arity: 1,
				Constructor: func(route Route, mux *http.ServeMux) {
					SetCmdHandler(mux, route)
				},
			},
			"proxy": {
				Arity: 1,
				Constructor: func(route Route, mux *http.ServeMux) {
					SetProxyHandler(mux, route)
				},
			},
		},
	}
}

func (env Environement) GetValue(name string) (Expr, bool) {
	for _, decl := range env.Declarations {
		if decl.key.lexeme == name {
			return decl.value, true
		}
	}

	return Expr{}, false
}
