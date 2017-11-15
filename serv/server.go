package serv

import (
	"log"
	"net/http"
	"path"
	"strings"
)

func CreateHandler(route Route) http.Handler {
	if IsGit(route) {
		rootPath, _ := GetRepoPath(route.Data)
		// return http.FileServer(http.Dir(path))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			filePath := strings.Replace(r.URL.String(), route.Path, "", 1)

			if filePath == "" {
				filePath = "index.html"
			}

			loc := path.Join(rootPath, filePath)
			log.Printf("serving %v from %v", r.URL.String(), loc)
			http.ServeFile(w, r, loc)
		})
	} else {
		log.Fatalf("I don't know what to do with %v", route)
		return nil
	}
}
