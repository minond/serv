package serv

import (
	"net/http"
	"strings"
)

type routeKind string

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

type RuntimeValue struct {
	Value string
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

const (
	cmdRoute      routeKind = "cmd"      // Wants a command string
	dirRoute      routeKind = "dir"      // Wants a directory
	gitRoute      routeKind = "git"      // Wants a git url
	proxyRoute    routeKind = "proxy"    // Wants url:port?
	redirectRoute routeKind = "redirect" // Wants a url
)

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
