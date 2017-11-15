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
	routeCmd       routeType = "cmd"       // Wants a command string
	routeDirectory routeType = "directory" // Warts direcotry
	routeGit       routeType = "git"       // Wants a git url
	routeProxy     routeType = "proxy"     // Wants url:port?
	routeRedirect  routeType = "redirect"  // Wants url
)

var (
	routeTypes = map[string]routeType{
		"cmd":       routeCmd,
		"directory": routeDirectory,
		"git":       routeGit,
		"proxy":     routeProxy,
		"redirect":  routeRedirect,
	}
)

func IsCmd(route Route) bool {
	return route.Type == routeCmd
}

func IsDirectory(route Route) bool {
	return route.Type == routeDirectory
}

func IsGit(route Route) bool {
	return route.Type == routeGit
}

func IsProxy(route Route) bool {
	return route.Type == routeProxy
}

func IsRedirect(route Route) bool {
	return route.Type == routeRedirect
}

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

		rpath := match[0][1]
		rdata := match[0][3]
		rtype, valid := routeTypes[match[0][2]]

		log.Printf("route match %v using %v to %v\n", rpath, rtype, rdata)

		if valid == false {
			panic(fmt.Sprintf("unknown route type: %v", match[0][2]))
		}

		routes = append(routes, Route{
			Path: rpath,
			Type: rtype,
			Data: rdata,
		})
	}

	return routes
}
