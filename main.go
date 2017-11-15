package main

import (
	"fmt"

	"github.com/minond/serv/server"
)

func main() {
	servfile, err := server.GetServfile()

	if err != nil {
		panic(fmt.Sprintf("error reading Servfile: %v", err))
	}

	routes := server.ParseServfile(servfile)

	fmt.Println(routes)
}
