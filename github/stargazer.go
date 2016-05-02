package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Stargazer struct {
	Timestamp string `json:"starred_at"`
}

func (s Stargazer) GetTimestamp() (int64, error) {
	t, err := time.Parse(time.RFC3339, s.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("An error occured while parsing the timestamp: %v", err)
	}
	return t.Unix(), nil
}

func GetStargazers(pageUrl, token string) (stargazers []Stargazer, link string, err error) {
	r, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return
	}

	r.Header.Add("Accept", "application/vnd.github.v3.star+json")
	r.Header.Add("Authorization", "token "+token)

	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Wrong error status while requesting stargazers: %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &stargazers)
	if err != nil {
		return
	}
	return stargazers, resp.Header.Get("Link"), nil
}
