package example

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evermax/stargraph/github"
)

// GetTimestamps gets the timestamps of the stars from the Github API
// count is the number of pages to query
// url is the Github API url of the repository you want to crawl
// token is the Github API token
func GetTimestamps(perPage int, url, token string) (timestamps []int64, err error) {
	if perPage > 0 {
		url = url + "?per_page=" + strconv.Itoa(perPage)
	}
	linkFormat := github.BuildLinksFormat(url)

	getParam := "?page="
	if strings.Contains(url, "?") {
		getParam = "&page="
	}
	var i = 1
	var last = 2
	var next int
	for {
		pageURL := url + getParam + strconv.Itoa(i)
		var stargazers []github.Stargazer
		var linkHeader string
		stargazers, linkHeader, err = github.GetStargazers(pageURL, token)
		if err != nil {
			return
		}

		var timestamp int64
		for _, star := range stargazers {
			timestamp, err = star.GetTimestamp()
			if err != nil {
				return
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
				err = fmt.Errorf("An error occured while parsing the header: %v, parser is %s, link header is %s", err, linkFormat, linkHeader)
				return
			}
		}

		if i == last {
			break
		}
		i++
	}
	return timestamps, nil
}
