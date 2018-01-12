package serv

import (
	"net/http"
	"strings"
)

// Translates an ast into setup and runtime informatin. `Match.expr` should be
// used to generate the `Server.Match` function and `Match.dcls` are the
// routes.
func Runtime(matches []Match) []Server {
	var servers []Server

	for _, match := range matches {
		var routes []Route

		Info("Generating %s", match.expr)

		for _, decl := range match.dcls {
			Info("Mounting %s", decl)

			switch decl.kind {
			case path:
				routes = append(routes, declToRoute(decl))

			default:
				Warn("Unknown declaration kind: %s", decl.kind)
			}
		}

		servers = append(servers, Server{
			Routes: routes,
			Match: func(r http.Request) bool {
				// XXX
				return false
			},
		})
	}

	return servers
}

func declToRoute(decl Declaration) Route {
	var kind routeKind
	var args []string

	switch decl.value.value.lexeme {
	case "cmd":
		kind = cmdRoute
	case "dir":
		kind = dirRoute
	case "git":
		kind = gitRoute
	case "proxy":
		kind = proxyRoute
	case "redirect":
		kind = redirectRoute
	default:
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
