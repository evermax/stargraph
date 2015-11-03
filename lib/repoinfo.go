package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Get the api url from a repo.
// The token is an API Github token to be able to lift of the 60 requests/hour limit
// The repo is a Github repo formated as follow `:username/:reponame`
func GetRepoInfo(token, repo string) (string, int, error) {
	url := "https://api.github.com/repos/" + repo
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	if token != "" {
		r.Header.Add("Authorization", "token "+token)
	}
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("Wrong status, expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	var info repoInfo
	if err := json.Unmarshal(bodyBytes, &info); err != nil {
		return "", 0, err
	}
	return "https://api.github.com/repositories/" + strconv.Itoa(info.ID) + "/stargazers", info.Count, nil
}

type repoInfo struct {
	ID    int `json:"id"`
	Count int `json:"stargazers_count"`
}
