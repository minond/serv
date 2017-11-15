package serv

import "testing"

func TestRepoPathGenerator(t *testing.T) {
	url := "https://github.com/minond/brainloller.git"
	path, err := GetRepoPath(url)

	if err != nil {
		t.Fatalf("error generating path: %v", err)
	} else if path != "repo/github.com/minond/brainloller.git" {
		t.Fatal("generated invalid path")
	}
}
