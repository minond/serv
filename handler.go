package main

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
func getRepoPath(repoURL string) (string, error) {
	ur, err := url.Parse(repoURL)

	if err != nil {
		return "", fmt.Errorf("error parsing url: %v", err)
	}

	return _path.Join(rootDir, ur.Hostname(), ur.EscapedPath()), nil
}

func pullGitRepoInterval(repoURL string) {
	info("pulling %v every %v", repoURL, *pullInterval)

	for {
		time.Sleep(*pullInterval)
		pullGitRepo(repoURL)
	}
}

func pullGitRepo(repoURL string) {
	path, err := getRepoPath(repoURL)

	if found, _ := fileExists(path); !found {
		return
	}

	info("running git pull on %v", path)

	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path

	if err = cmd.Run(); err != nil {
		warn("error running git pull on %v: %v", path, err)
	}
}

func checkoutGitRepo(repoURL string) (string, error) {
	path, err := getRepoPath(repoURL)

	if err != nil {
		return "", err
	}

	info("mkdir %v", path)
	err = os.MkdirAll(path, 0755)

	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", repoURL, path, "--depth=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return path, cmd.Run()
}

func localRepoExists(repoURL string) (bool, error) {
	path, err := getRepoPath(repoURL)

	if err != nil {
		return false, err
	}

	return fileExists(path)
}

func assertGitRepo(repoURL string) {
	if exists, _ := localRepoExists(repoURL); exists == false {
		if _, err := checkoutGitRepo(repoURL); err != nil {
			panic(fmt.Sprintf("error checking out git repo: %v", err))
		}
	}
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func assertDir(name string) {
	if exists, _ := fileExists(name); exists == false {
		panic(fmt.Sprintf("expecting %v directory which does not exists", name))
	}
}

func setProxyHandler(mux *http.ServeMux, route route) {
	proxyURL, err := url.Parse(route.data[0])
	proxyPath := proxyURL.Path

	if err != nil {
		panic(fmt.Sprintf("error parting proxy url (%v): %v", route.data[0], err))
	}

	proxy := func(w http.ResponseWriter, r *http.Request) {
		oldPath := r.URL.Path
		newPath := strings.Replace(oldPath, route.path, "", 1)

		r.URL.Path = proxyPath + newPath

		info("making request to %v", r.URL)
		handler := httputil.NewSingleHostReverseProxy(proxyURL)

		if r.Header.Get("Upgrade") != "websocket" {
			r.Header.Add("X-Forwarded-Proto", "https")
			r.Header.Add("X-Forwarded-Ssl", "on")
			r.Header.Add("X-Forwarded-Port", "443")
		}

		handler.ServeHTTP(w, r)
	}

	mux.HandleFunc(route.path, proxy)
	mux.HandleFunc(route.path+"/", proxy)
}

func setCmdHandler(mux *http.ServeMux, route route) {
	mux.HandleFunc(route.path, func(w http.ResponseWriter, r *http.Request) {
		parts := route.data
		cmd := exec.Command(parts[0], parts[1:]...)
		info("executing `%v` command", parts)

		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			warn("error running `%s` command: %v", strings.Join(parts, " "), err)
		}
	})
}

func setRedirectHandler(mux *http.ServeMux, route route) {
	mux.HandleFunc(route.path, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.data[0], http.StatusSeeOther)
	})
}

func setDirHandler(mux *http.ServeMux, route route) {
	serveFile := func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.Replace(r.URL.Path, route.path, "", 1)

		if filePath == "" {
			filePath = indexFile
		}

		loc := guessFileInDir(filePath, route.data[0])
		info("serving %v from %v", r.URL.String(), loc)
		http.ServeFile(w, r, loc)
	}

	slashRedirect := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, route.path+"/", http.StatusSeeOther)
	}

	if route.path == "/" {
		mux.HandleFunc(route.path, serveFile)
	} else {
		mux.HandleFunc(route.path, slashRedirect)
		mux.HandleFunc(route.path+"/", serveFile)
	}
}

func setGitHandler(mux *http.ServeMux, route route) {
	rootPath, _ := getRepoPath(route.data[0])

	if len(route.data) > 1 {
		route.data = []string{_path.Join(rootPath, route.data[1])}
	} else {
		route.data[0] = rootPath
	}

	setDirHandler(mux, route)
}

// NOTE This does have an issue in that if no local 404 file is found we should
// fallback to /404.html, but we don't since this function (or the handler)
// doesn't know about other routes and which one is on the / endpoint.
func guessFileInDir(file, dir string) string {
	origPath := _path.Join(dir, file)
	htmlPath := origPath + ".html"
	local404Path := _path.Join(dir, "404.html")

	if exists, _ := fileExists(htmlPath); exists == true {
		return htmlPath
	} else if exists, _ := fileExists(origPath); exists == true {
		return origPath
	} else if exists, _ := fileExists(local404Path); exists == true {
		return local404Path
	}

	return origPath
}
