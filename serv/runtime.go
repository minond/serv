package serv

import (
	"net/http"
	"strings"
)

type Environement struct {
	Matchers map[string]MatcherDef
}

type RuntimeValue struct {
	Value string
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

func NewEnvironment() Environement {
	return Environement{
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
	}
}
