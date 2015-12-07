package example

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evermax/stargraph/github"
)

// Get the timestamps of the stars from the Github API
// count is the number of pages to query
// url is the Github API url of the repository you want to crawl
// token is the Github API token
func GetTimestamps(perPage int, url, token string) ([]int64, error) {
	timestamps := make([]int64, 0)
	if perPage > 0 {
		url = url + "?per_page=" + strconv.Itoa(perPage)
	}
	linkFormat := github.BuildLinksFormat(url)

	getParam := "?page="
	if strings.Contains(url, "?") {
		getParam = "&page="
	}
	var i int = 1
	var last int = 2
	var next int
	for {
		pageUrl := url + getParam + strconv.Itoa(i)
		stargazers, linkHeader, err := github.GetStargazers(pageUrl, token)
		if err != nil {
			return make([]int64, 0), err
		}

		for _, star := range stargazers {
			timestamp, err := star.GetTimestamp()
			if err != nil {
				return make([]int64, 0), err
			}
			timestamps = append(timestamps, timestamp)
		}

		// If the header is empty, it is the only page
		if linkHeader == "" {
			break
		}
		// This is a little check because the last call will return a Link header
		// that doesn't have the same format
		if i < last {
			_, err = fmt.Sscanf(linkHeader, linkFormat, &next, &last)
			if err != nil {
				return make([]int64, 0), fmt.Errorf("An error occured while parsing the header: %v, parser is %s, link header is %s", err, linkFormat, linkHeader)
			}
		}

		if i == last {
			break
		}
		i++
	}
	return timestamps, nil
}
