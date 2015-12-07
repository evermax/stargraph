package github

import (
	"testing"
)

func TestGetRepoInfo(t *testing.T) {
	repo := "evermax/stargraph"
	expectedUrl := "https://api.github.com/repositories/45301830/stargazers"
	repoInfo, err := GetRepoInfo("", repo)
	if err != nil {
		t.Fatalf("An error occured when getting the repo informations: %v\n", err)
	}
	if !repoInfo.Exist() {
		t.Fatalf("The repository wasn't found on Github: %v\n", repoInfo)
	}

	if repoInfo.URL() != expectedUrl {
		t.Fatalf("The url gotten from Github API is %s, should be %s", repoInfo.URL(), expectedUrl)
	}
}

func TestExist(t *testing.T) {
	info := RepoInfo{}
	info.SetExist(true)
	if !info.Exist() {
		t.Fatalf("The exit variable wasn't changed.\n")
	}
}
