package github

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestGetRepoInfo(t *testing.T) {
	repo := "evermax/stargraph"
	expectedURL := "https://api.github.com/repositories/45301830/stargazers"
	repoInfo, err := GetRepoInfo("", repo)
	if err != nil {
		t.Fatalf("An error occured when getting the repo informations: %v\n", err)
	}
	if !repoInfo.Exist() {
		t.Fatalf("The repository wasn't found on Github: %v\n", repoInfo)
	}

	if repoInfo.URL() != expectedURL {
		t.Fatalf("The url gotten from Github API is %s, should be %s", repoInfo.URL(), expectedURL)
	}
}

func TestGetRepoInfo_WrongToken(t *testing.T) {
	repo := "evermax/stargraph"
	_, err := GetRepoInfo("qwerty", repo)
	if err == nil {
		t.Fatal("There should be an error because the token is incorrect")
	}
	if !strings.Contains(err.Error(), "Unexpected status") || !strings.Contains(err.Error(), strconv.Itoa(http.StatusUnauthorized)) {
		t.Fatalf("The error should be Unexpected status and the status should be %d: %s", http.StatusUnauthorized, err.Error())
	}
}

func TestExist(t *testing.T) {
	info := RepoInfo{}
	info.SetExist(true)
	if !info.Exist() {
		t.Fatalf("The exit variable wasn't changed.\n")
	}
}

func TestBuildFormatWithPerPage(t *testing.T) {
	var url = "test.com?perPage=3"
	var expectedFormat = "<" + url + "&page=%d>; rel=\"next\", <" + url + "&page=%d>; rel=\"last\""
	
	var result = BuildLinksFormat(url)
	if expectedFormat != result {
		t.Fatalf("expected following build format: %s\nGot: %s", expectedFormat, result)
	}
}

func TestBuildFormatWithoutPerPage(t *testing.T) {
	var url = "test.com"
	var expectedFormat = "<" + url + "?page=%d>; rel=\"next\", <" + url + "?page=%d>; rel=\"last\""
	
	var result = BuildLinksFormat(url)
	if expectedFormat != result {
		t.Fatalf("expected following build format: %s\nGot: %s", expectedFormat, result)
	}
}
