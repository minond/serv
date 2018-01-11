package main

// TODO domain checker???
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/minond/serv/parser"
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

	routeCmd      routeType = "cmd"      // Wants a command string
	routeDir      routeType = "dir"      // Wants a directory
	routeGit      routeType = "git"      // Wants a git url
	routeProxy    routeType = "proxy"    // Wants url:port?
	routeRedirect routeType = "redirect" // Wants a url
)

var (
	routeTypes = map[string]routeType{
		"cmd":      routeCmd,
		"dir":      routeDir,
		"git":      routeGit,
		"proxy":    routeProxy,
		"redirect": routeRedirect,
	}

	listen       = flag.String("listen", ":3002", "Host and port to listen on.")
	listenHttps  = flag.String("listenHttps", "", "Path to Let's Encrypt cache file instead of host/port.")
	config       = flag.String("config", "./Servfile", "Path to Servfile file.")
	pullInterval = flag.Duration("pullInterval", 15*time.Minute, "Interval git repos are pulled at.")
)

func IsCmd(route Route) bool {
	return route.Type == routeCmd
}

func IsDir(route Route) bool {
	return route.Type == routeDir
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

func ParseServfile(raw []byte) (routes []Route) {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	regex := regexp.MustCompile(`^([^\s|.]+)\s+([^\s|.]+)\s+(.+)$`)

	for _, line := range lines {
		clean := strings.TrimSpace(line)

		if clean == "" || clean[0] == byte('#') {
			continue
		}

		match := regex.FindAllStringSubmatch(clean, -1)

		if len(match) != 1 || len(match[0]) != 4 {
			log.Printf("ignoring configuration line: %v\n", line)
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
func GetRepoPath(repoURL string) (string, error) {
	ur, err := url.Parse(repoURL)

	if err != nil {
		return "", fmt.Errorf("error parsing url: %v", err)
	}

	return path.Join(rootDir, ur.Hostname(), ur.EscapedPath()), nil
}

func PullGitRepoInterval(repoURL string) {
	log.Printf("pulling %v every %v\n", repoURL, *pullInterval)
	for {
		time.Sleep(*pullInterval)
		PullGitRepo(repoURL)
	}
}

func PullGitRepo(repoURL string) {
	path, err := GetRepoPath(repoURL)

	if found, _ := FileExists(path); !found {
		return
	}

	log.Printf("running git pull on %v\n", path)

	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path

	if err = cmd.Run(); err != nil {
		log.Printf("error running git pull on %v: %v\n", path, err)
	}
}

// Clones repo into local folder
func CheckoutGitRepo(repoURL string) (string, error) {
	path, err := GetRepoPath(repoURL)

	if err != nil {
		return "", err
	}

	log.Printf("mkdir %v\n", path)
	err = os.MkdirAll(path, 0755)

	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", repoURL, path, "--depth=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return path, cmd.Run()
}

func LocalRepoExists(repoURL string) (bool, error) {
	path, err := GetRepoPath(repoURL)

	if err != nil {
		return false, err
	}

	return FileExists(path)
}

func AssertGitRepo(repoURL string) {
	if exists, _ := LocalRepoExists(repoURL); exists == false {
		if _, err := CheckoutGitRepo(repoURL); err != nil {
			panic(fmt.Sprintf("error checking out git repo: %v", err))
		}
	}
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func AssertDir(name string) {
	if exists, _ := FileExists(name); exists == false {
		panic(fmt.Sprintf("expecting %v directory which does not exists", name))
	}
}

func SetProxyHandler(mux *http.ServeMux, route Route) {
	proxyURL, err := url.Parse(route.Data)
	proxyPath := proxyURL.Path

	if err != nil {
		panic(fmt.Sprintf("error parting proxy url (%v): %v", route.Data, err))
	}

	proxy := func(w http.ResponseWriter, r *http.Request) {
		oldPath := r.URL.Path
		newPath := strings.Replace(oldPath, route.Path, "", 1)

		r.URL.Path = proxyPath + newPath

		log.Printf("making request to %v\n", r.URL)
		handler := httputil.NewSingleHostReverseProxy(proxyURL)
		handler.ServeHTTP(w, r)
	}

	mux.HandleFunc(route.Path, proxy)
	mux.HandleFunc(route.Path+"/", proxy)
}

func SetCmdHandler(mux *http.ServeMux, route Route) {
	mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(route.Data, " ")
		cmd := exec.Command(parts[0], parts[1:]...)
		log.Printf("executing `%v` command\n", parts)

		cmd.Stdout = w
		cmd.Stderr = w
		cmd.Run()
	})
}

func SetRedirectHandler(mux *http.ServeMux, route Route) {
	mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.Data, http.StatusSeeOther)
	})
}

func SetDirHandler(mux *http.ServeMux, route Route) {
	serveFile := func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.Replace(r.URL.Path, route.Path, "", 1)

		if filePath == "" {
			filePath = indexFile
		}

		loc := GuessFileInDir(filePath, route.Data)
		log.Printf("serving %v from %v\n", r.URL.String(), loc)
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
	route.Data = rootPath
	SetDirHandler(mux, route)
}

// NOTE This does have an issue in that if no local 404 file is found we should
// fallback to /404.html, but we don't since this function (or the handler)
// doesn't know about other routes and which one is on the / endpoint.
func GuessFileInDir(file, dir string) string {
	origPath := path.Join(dir, file)
	htmlPath := origPath + ".html"
	local404Path := path.Join(dir, "404.html")

	if exists, _ := FileExists(htmlPath); exists == true {
		return htmlPath
	} else if exists, _ := FileExists(origPath); exists == true {
		return origPath
	} else if exists, _ := FileExists(local404Path); exists == true {
		return local404Path
	} else {
		return origPath
	}
}

func main() {
	flag.Parse()

	fmt.Println(parser.Parse(`
case Host(dearme, minond, _) =>
  path / proxy(http://localhost:3002)

case Host(_, minond, co) =>
  path / redirect(https://minond.xyz)

# Don't allow anything at midnight
case DateTime(_, _, _, 0, 0, 0) =>
  path / pass(500)

case Host(_, minond, _) =>
  path /        git(https://github.com/minond/minond.github.io.git)
  path /servies git(https://github.com/minond/servies.git)
  path /static  dir(.)
  path /github  redirect(https://github.com/minond)
  path /ps      cmd(ps, aux)
  path /imdb    proxy(http://www.imdb.com:80)
  path /unibrow proxy(http://localhost:3001)

`))

	// fmt.Println(parser.Parse(`hi(1, 2, 3, 213) := xyz(x => x * 3)`))

	// log.Printf("reading configuration from %v\n", *config)
	// servfile, err := ioutil.ReadFile(*config)
	//
	// if err != nil {
	// 	panic(fmt.Sprintf("error reading Servfile: %v", err))
	// }
	//
	// routes := ParseServfile(servfile)
	// mux := http.NewServeMux()
	//
	// for _, route := range routes {
	// 	log.Printf("creating %v handler for %v\n", route.Type, route.Path)
	//
	// 	switch {
	// 	case IsGit(route):
	// 		AssertGitRepo(route.Data)
	// 		SetGitHandler(mux, route)
	// 		go PullGitRepoInterval(route.Data)
	//
	// 	case IsDir(route):
	// 		AssertDir(route.Data)
	// 		SetDirHandler(mux, route)
	//
	// 	case IsRedirect(route):
	// 		SetRedirectHandler(mux, route)
	//
	// 	case IsCmd(route):
	// 		SetCmdHandler(mux, route)
	//
	// 	case IsProxy(route):
	// 		SetProxyHandler(mux, route)
	//
	// 	default:
	// 		panic(fmt.Sprintf("invalid route type `%v` in %v\n", route.Type, route))
	// 	}
	// }
	//
	// if *listenHttps != "" {
	// 	log.Printf("starting https server on %v\n", *listen)
	//
	// 	var m letsencrypt.Manager
	//
	// 	if err := m.CacheFile(*listenHttps); err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	// Because of this:
	// 	// > Serve runs an HTTP/HTTPS web server using TLS certificates
	// 	// > obtained by the manager. The HTTP server redirects all requests to
	// 	// > the HTTPS server. The HTTPS server obtains TLS certificates as
	// 	// > needed and responds to requests by invoking http.DefaultServeMux.
	// 	http.DefaultServeMux = mux
	// 	log.Fatal(m.Serve())
	// } else {
	// 	log.Printf("starting http server on %v\n", *listen)
	// 	log.Fatal(http.ListenAndServe(*listen, mux))
	// }
}
