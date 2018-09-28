package main

import (
	"net/http"
)

// Runtime takes parsed declarations and matches and builds the working http
// handlers and an environment.
func runtime(decls []declaration, matches []match) ([]server, environement) {
	var servers []server
	env := newEnvironment(decls)

	for _, match := range matches {
		var routes []route

		info("Generating %s", match.expr)

		for _, decl := range match.decls {
			info("Mounting %s", decl)

			switch decl.kind {
			case path:
				routes = append(routes, declToRoute(env, decl))

			default:
				warn("Unknown declaration kind: %s", decl.kind)
			}
		}

		server := server{
			routes: routes,
			Match:  exprToMatch(env, match.expr),
			Mux:    buildMux(routes),
		}

		servers = append(servers, server)
	}

	return servers, env
}

func buildMux(routes []route) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range routes {
		info("Creating handler for %v", route.path)
		route.handler.constructor(route, mux)
	}

	return mux
}

func exprToMatch(env environement, expr expr) func(http.Request) bool {
	if expr.kind != call {
		fatal("Expecting a call but found %s instead", expr.kind)
	}

	var matcher matcher
	var args []string

	def, ok := env.matchers[expr.val.lexeme]

	if ok && def.arity != len(expr.args) {
		fatal("Wrong number of arguments for %s. Expected %d but got %d.",
			expr.val.lexeme, def.arity, len(expr.args))
		matcher = nullMatcher{}
	} else if !ok {
		warn("Unknown matcher kind: %s", expr.val.lexeme)
		matcher = nullMatcher{}
	} else {
		for _, arg := range expr.args {
			args = append(args, arg.lexeme)
		}

		matcher = def.constructor(args...)
	}

	return func(r http.Request) bool {
		return matcher.Match(r)
	}
}

func declToRoute(env environement, decl declaration) route {
	var args []string
	handler, ok := env.handlers[decl.val.val.lexeme]

	if !ok {
		fatal("Invalid route kind: %s",
			decl.val.val.lexeme)
	}

	for _, arg := range decl.val.args {
		args = append(args, arg.lexeme)
	}

	return route{
		handler: handler,
		path:    decl.key.lexeme,
		data:    args,
	}
}
