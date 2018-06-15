package serv

import (
	"net/http"
	"strings"
)

type environement struct {
	matchers     map[string]matcherDef
	handlers     map[string]handlerDef
	declarations []declaration
}

type runtimeValue struct {
	value string
}

type handlerDef struct {
	arity       int
	constructor func(route, *http.ServeMux)
}

type matcherDef struct {
	arity       int
	constructor func(...string) matcher
}

type matcher interface {
	Match(http.Request) bool
}

type nullMatcher struct{}

type hostMatcher struct {
	subdomain runtimeValue
	domain    runtimeValue
	tld       runtimeValue
}

func value(val string) runtimeValue {
	return runtimeValue{value: val}
}

func (v runtimeValue) equals(other string) bool {
	if v.value == "_" {
		return true
	}

	return v.value == other
}

func (n nullMatcher) Match(r http.Request) bool {
	return false
}

func (h hostMatcher) Match(r http.Request) bool {
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

	return h.subdomain.equals(subdomain) &&
		h.domain.equals(domain) &&
		h.tld.equals(tld)
}

func newEnvironment(decls []declaration) environement {
	return environement{
		declarations: decls,
		matchers: map[string]matcherDef{
			"Host": {
				arity: 3,
				constructor: func(args ...string) matcher {
					return hostMatcher{
						subdomain: value(args[0]),
						domain:    value(args[1]),
						tld:       value(args[2]),
					}
				},
			},
		},

		handlers: map[string]handlerDef{
			"git": {
				arity: 1,
				constructor: func(route route, mux *http.ServeMux) {
					assertGitRepo(route.data[0])
					setGitHandler(mux, route)
					go pullGitRepoInterval(route.data[0])
				},
			},
			"dir": {
				arity: 1,
				constructor: func(route route, mux *http.ServeMux) {
					assertDir(route.data[0])
					setDirHandler(mux, route)
				},
			},
			"redirect": {
				arity: 1,
				constructor: func(route route, mux *http.ServeMux) {
					setRedirectHandler(mux, route)
				},
			},
			"cmd": {
				arity: 1,
				constructor: func(route route, mux *http.ServeMux) {
					setCmdHandler(mux, route)
				},
			},
			"proxy": {
				arity: 1,
				constructor: func(route route, mux *http.ServeMux) {
					setProxyHandler(mux, route)
				},
			},
		},
	}
}

func (env environement) GetValue(name string) (expr, bool) {
	for _, decl := range env.declarations {
		if decl.key.lexeme == name {
			return decl.val, true
		}
	}

	return expr{}, false
}
