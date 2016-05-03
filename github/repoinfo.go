package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// IRepoInfo is an interface for the RepoInfo for test purposes 
// it returns the star count and the URL for the repository.
// It might evolve over time if some other fields are needed.
type IRepoInfo interface {
	StarCount() int
	URL() string
}

const (
	// GithubRepoURL is the base URL to get the data for a repository
	GithubRepoURL = "https://api.github.com/repos/"
	// GithubStarURLFormat will help build the URL to request the stars.
	// It will use the id of the repository to make the query.
	GithubStarURLFormat = "https://api.github.com/repositories/%d/stargazers"
)

// RepoInfo is the entity that will be put in the database but
// it is also used to parse the response from the Github API. 
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

// URL will return the URL to request the stars for the repository
func (info RepoInfo) URL() string {
	return fmt.Sprintf(GithubStarURLFormat, info.ID)
}

// StarCount will return the number of stars for the repository
func (info RepoInfo) StarCount() int {
	return info.Count
}

// Exist flag is not in the database, it is just to know whether
// it needs to be created in the database or not.
func (info RepoInfo) Exist() bool {
	return info.exist
}

// SetExist Is used to set the flag to the passed argument
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

// BuildLinksFormat will build the Link header parser.
// See https://developer.github.com/v3/activity/starring/
func BuildLinksFormat(url string) string {
	if strings.Contains(url, "?") {
		return "<" + url + "&page=%d>; rel=\"next\", <" + url + "&page=%d>; rel=\"last\""
	}
	return "<" + url + "?page=%d>; rel=\"next\", <" + url + "?page=%d>; rel=\"last\""
}
