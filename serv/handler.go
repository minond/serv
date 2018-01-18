package serv

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	_path "path"
	"strings"
	"time"
)

const (
	indexFile = "index.html"
	rootDir   = "repo"
)

var (
	pullInterval = flag.Duration("pullInterval", 15*time.Minute, "Interval git repos are pulled at.")
)

// Turns https://github.com/minond/minond.github.io.git into
// repo/github.com/minond/minond.github.io.git
func GetRepoPath(repoURL string) (string, error) {
	ur, err := url.Parse(repoURL)

	if err != nil {
		return "", fmt.Errorf("error parsing url: %v", err)
	}

	return _path.Join(rootDir, ur.Hostname(), ur.EscapedPath()), nil
}

func PullGitRepoInterval(repoURL string) {
	Info("Pulling %v every %v", repoURL, *pullInterval)

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

	Info("Running git pull on %v", path)

	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path

	if err = cmd.Run(); err != nil {
		Fatal("Error running git pull on %v: %v", path, err)
	}
}

// Clones repo into local folder
func CheckoutGitRepo(repoURL string) (string, error) {
	path, err := GetRepoPath(repoURL)

	if err != nil {
		return "", err
	}

	Info("Mkdir %v", path)
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

		Info("Making request to %v", r.URL)
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
		Info("Executing `%v` command", parts)

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
		Info("Serving %v from %v", r.URL.String(), loc)
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
	origPath := _path.Join(dir, file)
	htmlPath := origPath + ".html"
	local404Path := _path.Join(dir, "404.html")

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
