package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/evermax/stargraph/lib"
)

var (
	repo, token string
	batch       int
)

func init() {
	flag.StringVar(&repo, "r", "evermax/stargraph", "Github Project repository using format :username/:repo. Default: evermax/stargraph")
	flag.StringVar(&token, "t", "", "Github API token\nYou can go on to the following link to know how to get one: https://github.com/blog/1509-personal-api-tokens")
	flag.IntVar(&batch, "n", 100, "Number of stars per request. Default: 100")
}

func main() {
	flag.Parse()

	fmt.Printf("Starting github star graph of %s\n", repo)
	startDate := time.Now()
	repoUrl, starCount, err := lib.GetRepoInfo(token, repo)
	//repoUrl, _, err := lib.GetRepoInfo(token, repo)
	if err != nil {
		fmt.Printf("An error occured while getting the repository info: %v\n", err)
		return
	}
	//timestamps, err := lib.GetTimestamps(batch, repoUrl, token)
	timestamps, err := lib.GetTimestampsDistributed(starCount, batch, repoUrl, token)
	if err != nil {
		fmt.Printf("An error occured while getting the stars from Github: %v\n", err)
		return
	}
	endDate := time.Now()
	duration := endDate.Sub(startDate)
	fmt.Printf("Timestamps delivered in %v\n", duration)
	fmt.Printf("Timestamps: %v\n", timestamps)
}
