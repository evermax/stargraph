package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// For test purpose
// May evolve
type IRepoInfo interface {
	StarCount() int
	URL() string
}

const (
	GithubRepoURL = "https://api.github.com/repos/"
)

type RepoInfo struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Count        int     `json:"stargazers_count"`
	CreationDate string  `json:"created_at"`
	LastStarDate string  `json:"last_star_date,omitempty"`
	LastUpdate   string  `json:"last_update,omitempty"`
	WorkedOn     bool    `json:"worked_on,omitempty"`
	Timestamps   []int64 `json:"timestamps,omitempty"`
	exist        bool
}

func (info RepoInfo) URL() string {
	return "https://api.github.com/repositories/" + strconv.Itoa(info.ID) + "/stargazers"
}

func (info RepoInfo) StarCount() int {
	return info.Count
}

func (info RepoInfo) Exist() bool {
	return info.exist
}

func (info *RepoInfo) SetExist(exist bool) {
	info.exist = exist
}

// GetRepoInfo get the api url from a repo.
// The token is an API Github token to be able to lift off the 60 requests/hour limit
// The repo is a Github repo formated as follow `:username/:reponame`
func GetRepoInfo(token, repo string) (info RepoInfo, err error) {
	url := GithubRepoURL + repo
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if token != "" {
		r.Header.Add("Authorization", "token "+token)
	}

	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return
	}

	// The repo doesn't exist so no error, just empty repo
	if resp.StatusCode == http.StatusNotFound {
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Unexpected status, expected %d, got %d\nThe body was: %s\n", http.StatusOK, resp.StatusCode, string(bodyBytes))
		return
	}

	if err = json.Unmarshal(bodyBytes, &info); err != nil {
		return
	}
	info.exist = true

	return info, nil
}

func BuildLinksFormat(url string) string {
	if strings.Contains(url, "?") {
		return "<" + url + "&page=%d>; rel=\"next\", <" + url + "&page=%d>; rel=\"last\""
	}
	return "<" + url + "?page=%d>; rel=\"next\", <" + url + "?page=%d>; rel=\"last\""
}
