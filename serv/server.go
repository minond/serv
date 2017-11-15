package serv

import (
	"log"
	"net/http"
	"path"
	"strings"
)

const (
	indexFile = "index.html"
)

func SetDirectoryHandler(mux *http.ServeMux, route Route) {
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
