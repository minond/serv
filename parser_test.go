package serv

import (
	"io/ioutil"
	"log"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func checkRouteCount(t *testing.T, routes []Route, expected int) {
	if len(routes) != expected {
		t.Fatal("should come back as only one route")
	}
}

func checkRouteParts(t *testing.T, route Route, expected Route) {
	if route.Path != expected.Path {
		t.Fatal("invalid path")
	}

	if route.Type != expected.Type {
		t.Fatal("invalid type")
	}

	if route.Data != expected.Data {
		t.Fatal("invalid data")
	}
}

func TestGitRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeGit}

	if IsGit(route) == false {
		t.Fatal("this should come back as a git route")
	}

	if IsProxy(route) == true {
		t.Fatal("this should not come back as a proxy route")
	}
}

func TestProxyRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeProxy}

	if IsProxy(route) == false {
		t.Fatal("this should come back as a proxy route")
	}

	if IsGit(route) == true {
		t.Fatal("this should not come back as a git route")
	}
}

func TestParsesSingleLine(t *testing.T) {
	raw := `/ git https://gh.com/path/to/repo.git`
	routes := ParseServfile([]byte(raw))

	checkRouteCount(t, routes, 1)

	checkRouteParts(t, routes[0], Route{
		Path: "/",
		Type: routeGit,
		Data: "https://gh.com/path/to/repo.git",
	})
}

func TestParsesSingleWeirdLine(t *testing.T) {
	raw := `  /                 git      https://gh.com/path/to/repo.git     `
	routes := ParseServfile([]byte(raw))

	checkRouteCount(t, routes, 1)

	checkRouteParts(t, routes[0], Route{
		Path: "/",
		Type: routeGit,
		Data: "https://gh.com/path/to/repo.git",
	})
}

func TestParsesMultipleLines(t *testing.T) {
	raw := `
/one git https://gh.com/path/to/repo-one.git
/two git https://gh.com/path/to/repo-two.git
`

	routes := ParseServfile([]byte(raw))

	checkRouteCount(t, routes, 2)

	checkRouteParts(t, routes[0], Route{
		Path: "/one",
		Type: routeGit,
		Data: "https://gh.com/path/to/repo-one.git",
	})

	checkRouteParts(t, routes[1], Route{
		Path: "/two",
		Type: routeGit,
		Data: "https://gh.com/path/to/repo-two.git",
	})
}
