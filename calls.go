package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func getRepoInfo(token, repo string) (string, int, error) {
	url := "https://api.github.com/repos/" + repo
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := passRequest(r, token)
	if err != nil {
		return "", 0, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	var repoInfo githubRepoInfo
	if err := json.Unmarshal(bodyBytes, &repoInfo); err != nil {
		return "", 0, err
	}
	return "https://api.github.com/repositories/" + strconv.Itoa(repoInfo.Id) + "/stargazers", repoInfo.Count, nil
}

func getDataFromUrl(url string, token string, stargazers *[]stargazer, position *int, timestamps *[]int64) (string, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("An error occured when creating the request with url: %s\n", url)
	}
	r.Header.Add("Accept", "application/vnd.github.v3.star+json")
	resp, err := passRequest(r, token)
	if err != nil {
		return "", err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error parsing the response body %v\nBody was %s", err, string(bodyBytes))
	}
	var stars []stargazer
	if err := json.Unmarshal(bodyBytes, &stars); err != nil {
		return "", fmt.Errorf("An error occured when parsing the response from Github %v\n", err)
	}
	for _, star := range stars {
		star.PageUrl = url
		star.Count = *position + 1
		t, err := time.Parse(time.RFC3339, star.Timestamp)
		if err != nil {
			return "", err
		}
		timestamp := t.Unix() * 1000
		var pos = *position
		(*stargazers)[pos] = star
		(*timestamps)[pos] = timestamp
		*position = pos + 1
	}
	return resp.Header.Get("Link"), nil
}

func passRequest(r *http.Request, token string) (*http.Response, error) {
	if token != "" {
		r.Header.Add("Authorization", "token "+token)
	}
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wrong status, expected %d, got %d\n", http.StatusOK, resp.StatusCode)
	}
	return resp, nil
}

type githubRepoInfo struct {
	Id    int `json:"id"`
	Count int `json:"stargazers_count"`
}
