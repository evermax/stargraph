package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/evermax/stargraph/lib"
)

var (
	repo, token string
	batch       int
	concurrent  bool
)

func init() {
	flag.StringVar(&repo, "r", "evermax/stargraph", "Github Project repository using format :username/:repo. Default: evermax/stargraph")
	flag.StringVar(&token, "t", "", "Github API token\nYou can go on to the following link to know how to get one: https://github.com/blog/1509-personal-api-tokens")
	flag.IntVar(&batch, "n", 100, "Number of stars per request. Default: 100")
	flag.BoolVar(&concurrent, "c", true, "Whether you want to run the requests concurrently or not. Default: true")
}

func main() {
	flag.Parse()

	fmt.Printf("Starting github star graph of %s\n", repo)
	startDate := time.Now()
	repoUrl, starCount, err := lib.GetRepoInfo(token, repo)
	if err != nil {
		fmt.Printf("An error occured while getting the repository info: %v\n", err)
		return
	}
	var timestamps []int64
	if concurrent {
		timestamps, err = lib.GetTimestampsDistributed(starCount, batch, repoUrl, token)
	} else {
		timestamps, err = lib.GetTimestamps(batch, repoUrl, token)
	}
	if err != nil {
		fmt.Printf("An error occured while getting the stars from Github: %v\n", err)
		return
	}
	endDate := time.Now()
	duration := endDate.Sub(startDate)
	fmt.Printf("Timestamps gotten in %v\n", duration)

	fmt.Println("Persisting them into canvas format...")
	canvasFile, err := os.Create("canvasdb.json")
	if err != nil {
		fmt.Printf("An error occured when creating the canvas file %v\n", err)
		return
	}
	defer canvasFile.Close()
	if err = lib.WriteCanvasJS(timestamps, canvasFile); err != nil {
		fmt.Printf("An error occured when writing the canvas file%v\n", err)
	}
	fmt.Println("Done.")

	fmt.Println("Persisting them into jqplot format...")
	jqplotFile, err := os.Create("jqplotdb.json")
	if err != nil {
		fmt.Printf("An error occured when creating the jqplot file %v\n", err)
		return
	}
	defer jqplotFile.Close()
	if err = lib.WriteCanvasJS(timestamps, jqplotFile); err != nil {
		fmt.Printf("An error occured when writing the jqplot file %v\n", err)
		return
	}
	fmt.Println("Done.")

	fmt.Println("Drawing the graph image...")
	graphFile, err := os.Create("graph.png")
	if err != nil {
		fmt.Printf("An error occured when creating the graph image %v\n", err)
		return
	}
	defer graphFile.Close()
	if err = lib.PlotGraph("Graph of "+repo, timestamps, graphFile); err != nil {
		fmt.Printf("An error occured when drawing the imge%v\n", err)
		return
	}
	fmt.Println("Done.")
}
