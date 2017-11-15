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
			if exists, _ := serv.LocalRepoExists(route.Data); exists == false {
				if _, err = serv.CheckoutGitRepo(route.Data); err != nil {
					panic(fmt.Sprintf("error checking out git repo: %v", err))
				}
			}

			serv.SetGitHandler(mux, route)
		} else if serv.IsDirectory(route) {
			serv.SetDirectoryHandler(mux, route)
		}
	}

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":3002", mux))
}
