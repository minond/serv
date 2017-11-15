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

	for _, route := range routes {
		if serv.IsGit(route) {
			exists, _ := serv.LocalRepoExists(route.Data)

			if exists == false {
				_, _ = serv.CheckoutGitRepo(route.Data)
			}
		}

		log.Printf("creating handler for %v", route.Path)
		http.Handle(route.Path, serv.CreateHandler(route))
	}

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":3002", nil))
}
