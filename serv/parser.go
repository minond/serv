package serv

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

type routeType string

type Route struct {
	// Top-level url path this route should match to
	Path string

	// See `routeType`
	Type routeType

	// Information passed to prep function. Type of data depends on Route.Type
	Data string
}

const (
	routeGit   routeType = "git"   // Wants a git url
	routeProxy routeType = "proxy" // Wants url:port
)

func GetServfile() ([]byte, error) {
	return ioutil.ReadFile("./Servfile")
}

func ParseServfile(raw []byte) (routes []Route) {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	regex := regexp.MustCompile(`^([^\s|.]+)\s+([^\s|.]+)\s+(.+)$`)

	for _, line := range lines {
		match := regex.FindAllStringSubmatch(strings.TrimSpace(line), -1)

		if len(match) != 1 || len(match[0]) != 4 {
			log.Printf("ignoring configuration line: %v", line)
			continue
		}

		log.Printf("route match: `%v` using %v to %v\n", match[0][1], match[0][2], match[0][3])

		switch match[0][2] {
		case "git":
			routes = append(routes, Route{
				Path: match[0][1],
				Type: routeGit,
				Data: match[0][3],
			})

		default:
			panic(fmt.Sprintf("unknown route type: %v", match[0][2]))
		}
	}

	return routes
}

func IsProxy(route Route) bool {
	return route.Type == routeProxy
}

func IsGit(route Route) bool {
	return route.Type == routeGit
}
