package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type routeType string

type Route struct {
	Path string
	Type routeType
	Data string
}

const (
	indexFile = "index.html"
	rootDir   = "repo"

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

// Turns https://github.com/minond/minond.github.io.git into
// repo/github.com/minond/minond.github.io.git
func GetRepoPath(repoUrl string) (string, error) {
	ur, err := url.Parse(repoUrl)

	if err != nil {
		return "", fmt.Errorf("error parsing url: %v", err)
	}

	return path.Join(rootDir, ur.Hostname(), ur.EscapedPath()), nil
}

// Clones repo into local folder
func CheckoutGitRepo(repoUrl string) (string, error) {
	path, err := GetRepoPath(repoUrl)

	if err != nil {
		return "", err
	}

	log.Printf("mkdir %v\n", path)
	err = os.MkdirAll(path, 0755)

	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", repoUrl, path, "--depth=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return path, cmd.Run()
}

func LocalRepoExists(repoUrl string) (bool, error) {
	path, err := GetRepoPath(repoUrl)

	if err != nil {
		return false, err
	}

	return DirectoryExists(path)
}

func AssertGitRepo(repoUrl string) {
	if exists, _ := LocalRepoExists(repoUrl); exists == false {
		if _, err := CheckoutGitRepo(repoUrl); err != nil {
			panic(fmt.Sprintf("error checking out git repo: %v", err))
		}
	}
}

func DirectoryExists(name string) (bool, error) {
	_, err := os.Stat(name)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func AssertDirectory(name string) {
	if exists, _ := DirectoryExists(name); exists == false {
		panic(fmt.Sprintf("expecting %v directory which does not exists", name))
	}
}

func SetRedirectHandler(mux *http.ServeMux, route Route) {
	mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.Data, http.StatusSeeOther)
	})
}

func SetDirectoryHandler(mux *http.ServeMux, route Route) {
	serveFile := func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.Replace(r.URL.String(), route.Path, "", 1)

		if filePath == "" {
			filePath = indexFile
		}

		loc := path.Join(route.Data, filePath)
		log.Printf("serving %v from %v", r.URL.String(), loc)
		http.ServeFile(w, r, loc)
	}

	slashRedirect := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.Path+"/", http.StatusSeeOther)
	}

	if route.Path == "/" {
		mux.HandleFunc(route.Path, serveFile)
	} else {
		mux.HandleFunc(route.Path, slashRedirect)
		mux.HandleFunc(route.Path+"/", serveFile)
	}
}

func SetGitHandler(mux *http.ServeMux, route Route) {
	rootPath, _ := GetRepoPath(route.Data)

	serveFile := func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.Replace(r.URL.String(), route.Path, "", 1)

		if filePath == "" {
			filePath = indexFile
		}

		loc := path.Join(rootPath, filePath)
		log.Printf("serving %v from %v", r.URL.String(), loc)
		http.ServeFile(w, r, loc)
	}

	slashRedirect := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.Path+"/", http.StatusSeeOther)
	}

	if route.Path == "/" {
		mux.HandleFunc(route.Path, serveFile)
	} else {
		mux.HandleFunc(route.Path, slashRedirect)
		mux.HandleFunc(route.Path+"/", serveFile)
	}
}

func main() {
	servfile, err := GetServfile()

	if err != nil {
		panic(fmt.Sprintf("error reading Servfile: %v", err))
	}

	routes := ParseServfile(servfile)
	mux := http.NewServeMux()

	for _, route := range routes {
		log.Printf("creating %v handler for %v", route.Type, route.Path)

		if IsGit(route) {
			AssertGitRepo(route.Data)
			SetGitHandler(mux, route)
		} else if IsDirectory(route) {
			AssertDirectory(route.Data)
			SetDirectoryHandler(mux, route)
		} else if IsRedirect(route) {
			SetRedirectHandler(mux, route)
		}
	}

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":3002", mux))
}
