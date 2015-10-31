package main

import (
	"flag"
	"fmt"
	"strconv"
)

var (
	project, token string
	batch          int
)

func init() {
	flag.StringVar(&project, "p", "evermax/stargraph", "Github Projet")
	flag.StringVar(&token, "t", "", "Github API token")
	flag.IntVar(&batch, "n", 100, "Number of stars per request")
}

func main() {
	flag.Parse()

	fmt.Printf("Starting github start graph of %s\n", project)
	repoUrl, count, err := getRepoInfo(token, project)
	if err != nil {
		fmt.Printf("An error occured while getting the info about the repo: %v\n", err)
		return
	}

	repoUrl = repoUrl + "?per_page=" + strconv.Itoa(batch)
	url := repoUrl + "&page=1"
	linkFormat := "<" + repoUrl + "&page=%d>; rel=\"next\", <" +
		repoUrl + "&page=%d>; rel=\"last\""

	var position int = 0
	var stargazers = make([]stargazer, count)
	var timestamps = make([]int64, count)
	links, err := getDataFromUrl(url, token, &stargazers, &position, &timestamps)
	if err != nil {
		fmt.Printf("An error occured while getting data from github: %v\n", err)
		return
	}

	for {
		if links == "" {
			fmt.Printf("Link header not found in the response, aborting.\n")
			break
		}
		var next, last int
		_, err := fmt.Sscanf(links, linkFormat, &next, &last)
		if err != nil {
			fmt.Printf("An error occured while parsing the Links header: %v\n"+
				"Using this format: %s\nThe content was: %s\n", err, linkFormat, links)
			return
		}
		nextUrl := repoUrl + "&page=" + strconv.Itoa(next)
		links, err = getDataFromUrl(nextUrl, token, &stargazers, &position, &timestamps)
		if err != nil {
			fmt.Printf("The following error occured with the url: %s\n%v\n", nextUrl, err)
			return
		}
		if next == last {
			break
		}
	}

	if err := persistData(timestamps); err != nil {
		fmt.Printf("An error occured while persisting the data: %v", err)
	}
	if err := plotGraph(timestamps); err != nil {
		fmt.Printf("An error occured while plotting the graph: %v", err)
		return
	}
}

type stargazer struct {
	Timestamp string `json:"starred_at"`
	User      user   `json:"user"`
	Count     int    `json:"count"`
	PageUrl   string `json:"pageUrl"`
}

type user struct {
	Id int `json:"id"`
}
