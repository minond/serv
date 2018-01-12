package serv

import "net/http"

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

const (
	cmdRoute      routeKind = "cmd"      // Wants a command string
	dirRoute      routeKind = "dir"      // Wants a directory
	gitRoute      routeKind = "git"      // Wants a git url
	proxyRoute    routeKind = "proxy"    // Wants url:port?
	redirectRoute routeKind = "redirect" // Wants a url
)
