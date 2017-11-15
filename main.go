package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minond/serv/serv"
)

func main() {
	servfile, err := serv.GetServfile()

	if err != nil {
		panic(fmt.Sprintf("error reading Servfile: %v", err))
	}

	routes := serv.ParseServfile(servfile)
	mux := http.NewServeMux()

	for _, route := range routes {
		log.Printf("creating handler for %v", route.Path)

		if serv.IsGit(route) {
			serv.AssertGitRepo(route.Data)
			serv.SetGitHandler(mux, route)
		} else if serv.IsDirectory(route) {
			serv.AssertDirectory(route.Data)
			serv.SetDirectoryHandler(mux, route)
		}
	}

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":3002", mux))
}
