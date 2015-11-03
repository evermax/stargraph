package lib

import (
	"testing"
)

func TestGetRepoInfo(t *testing.T) {
	repo := "evermax/stargraph"
	expectedUrl := "https://api.github.com/repositories/45301830/stargazers"
	url, _, err := GetRepoInfo("", repo)
	if err != nil {
		t.Fatalf("An error occured when getting the repo informations: %v\n", err)
	}
	if url != expectedUrl {
		t.Fatalf("The url gotten from Github API is %s, should be %s", url, expectedUrl)
	}
}
