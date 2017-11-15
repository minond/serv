package serv

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
)

const (
	rootDir = "repo"
)

// Turns https://github.com/minond/minond.github.io.git into
// repo/github.com/minond/minond.github.io.git
func GetRepoPath(repoUrl string) (string, error) {
	ur, err := url.Parse(repoUrl)

	if err != nil {
		return "", fmt.Errorf("error parsing url: %v", err)
	}

	return path.Join(rootDir, ur.Hostname(), ur.EscapedPath()), nil
}

// Clones repo into local folder
func CheckoutGitRepo(repoUrl string) (string, error) {
	path, err := GetRepoPath(repoUrl)

	if err != nil {
		return "", err
	}

	log.Printf("mkdir %v\n", path)
	err = os.MkdirAll(path, 0777)

	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", repoUrl, path, "--depth=1")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return path, cmd.Run()
}

func LocalRepoExists(repoUrl string) (bool, error) {
	path, err := GetRepoPath(repoUrl)

	if err != nil {
		return false, err
	}

	return DirectoryExists(path)
}

func AssertGitRepo(repoUrl string) {
	if exists, _ := LocalRepoExists(repoUrl); exists == false {
		if _, err := CheckoutGitRepo(repoUrl); err != nil {
			panic(fmt.Sprintf("error checking out git repo: %v", err))
		}
	}
}
