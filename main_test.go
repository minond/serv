package main

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

func TestCmdRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeCmd}

	if IsCmd(route) == false {
		t.Fatal("this should come back as a cmd route")
	}
}

func TestDirectoryRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeDir}

	if IsDir(route) == false {
		t.Fatal("this should come back as a directory route")
	}
}

func TestGitRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeGit}

	if IsGit(route) == false {
		t.Fatal("this should come back as a git route")
	}
}

func TestProxyRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeProxy}

	if IsProxy(route) == false {
		t.Fatal("this should come back as a proxy route")
	}
}

func TestRedirectRouteTypeChecker(t *testing.T) {
	route := Route{Type: routeRedirect}

	if IsRedirect(route) == false {
		t.Fatal("this should come back as a redirect route")
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

func TestParsesAllTypes(t *testing.T) {
	raw := `
/one cmd who
/two dir .
/three git https://gh.com/path/to/repo.git
/four proxy localhost:3001
/five redirect http://google.com
`

	routes := ParseServfile([]byte(raw))

	checkRouteCount(t, routes, 5)

	checkRouteParts(t, routes[0], Route{
		Path: "/one",
		Type: routeCmd,
		Data: "who",
	})

	checkRouteParts(t, routes[1], Route{
		Path: "/two",
		Type: routeDir,
		Data: ".",
	})

	checkRouteParts(t, routes[2], Route{
		Path: "/three",
		Type: routeGit,
		Data: "https://gh.com/path/to/repo.git",
	})

	checkRouteParts(t, routes[3], Route{
		Path: "/four",
		Type: routeProxy,
		Data: "localhost:3001",
	})

	checkRouteParts(t, routes[4], Route{
		Path: "/five",
		Type: routeRedirect,
		Data: "http://google.com",
	})
}

func TestRepoPathGenerator(t *testing.T) {
	url := "https://github.com/minond/brainloller.git"
	path, err := GetRepoPath(url)

	if err != nil {
		t.Fatalf("error generating path: %v", err)
	} else if path != "repo/github.com/minond/brainloller.git" {
		t.Fatal("generated invalid path")
	}
}

func TestHostCleanerFunction(t *testing.T) {
	var perms = []struct{ raw, clean string }{
		{"testing:80", "testing"},
		{"www.testing:80", "www.testing"},
		{"www.testing", "www.testing"},
		{"localhost:3432", "localhost"},
		{"localhost", "localhost"},
	}

	for _, perm := range perms {
		if ret := GetCleanHost(perm.raw); ret != perm.clean {
			t.Fatalf("expected %v but got %v\n", perm.clean, ret)
		}
	}
}
