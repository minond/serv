package serv

import (
	"net/http"
)

// Runtime takes parsed declarations and matches and builds the working http
// handlers and an environment.
func Runtime(decls []declaration, matches []match) ([]server, environement) {
	var servers []server
	env := newEnvironment(decls)

	for _, match := range matches {
		var routes []route

		Info("Generating %s", match.expr)

		for _, decl := range match.decls {
			Info("Mounting %s", decl)

			switch decl.kind {
			case path:
				routes = append(routes, declToRoute(env, decl))

			default:
				Warn("Unknown declaration kind: %s", decl.kind)
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
		Info("Creating handler for %v", route.path)
		route.handler.constructor(route, mux)
	}

	return mux
}

func exprToMatch(env environement, expr expr) func(http.Request) bool {
	if expr.kind != call {
		Fatal("Expecting a call but found %s instead", expr.kind)
	}

	var matcher matcher
	var args []string

	def, ok := env.matchers[expr.val.lexeme]

	if ok && def.arity != len(expr.args) {
		Fatal("Wrong number of arguments for %s. Expected %d but got %d.",
			expr.val.lexeme, def.arity, len(expr.args))
		matcher = nullMatcher{}
	} else if !ok {
		Warn("Unknown matcher kind: %s", expr.val.lexeme)
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
		Fatal("Invalid route kind: %s",
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
