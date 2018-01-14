package serv

import (
	"net/http"
	"strings"
)

// Translates an ast into setup and runtime informatin. `Match.expr` should be
// used to generate the `Server.Match` function and `Match.decls` are the
// routes.
func Runtime(decls []Declaration, matches []Match) ([]Server, Environement) {
	var servers []Server
	env := NewEnvironment(decls)

	for _, match := range matches {
		var routes []Route

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

		servers = append(servers, Server{
			Routes: routes,
			Match:  exprToMatch(env, match.expr),
		})
	}

	return servers, env
}

func exprToMatch(env Environement, expr Expr) func(http.Request) bool {
	if expr.kind != call {
		Fatal("Expecting a call but found %s instead", expr.kind)
	}

	var matcher Matcher
	var args []string

	def, ok := env.Matchers[expr.value.lexeme]

	if ok && def.Arity != len(expr.args) {
		Fatal("Wrong number of arguments for %s. Expected %d but got %d.",
			expr.value.lexeme, def.Arity, len(expr.args))
		matcher = NullMatcher{}
	} else if !ok {
		Warn("Unknown matcher kind: %s", expr.value.lexeme)
		matcher = NullMatcher{}
	} else {
		for _, arg := range expr.args {
			args = append(args, arg.lexeme)
		}

		matcher = def.Constructor(args...)
	}

	return func(r http.Request) bool {
		return matcher.Match(r)
	}
}

func declToRoute(env Environement, decl Declaration) Route {
	var args []string
	kind, ok := env.RouteKinds[decl.value.value.lexeme]

	if !ok {
		Fatal("Invalid route kind: %s",
			decl.value.value.lexeme)
	}

	for _, arg := range decl.value.args {
		args = append(args, arg.lexeme)
	}

	return Route{
		Kind: kind,
		Path: decl.key.lexeme,
		Data: strings.Join(args, " "),
	}
}
