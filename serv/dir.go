package serv

import (
	"fmt"
	"os"
)

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
