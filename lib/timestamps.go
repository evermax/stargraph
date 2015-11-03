package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Get the timestamps of the stars from the Github API
// count is the number of pages to query
// url is the Github API url of the repository you want to crawl
// token is the Github API token
func GetTimestamps(batch int, url, token string) ([]int64, error) {
	timestamps := make([]int64, 0)
	if batch > 0 {
		url = url + "?per_page=" + strconv.Itoa(batch)
	}
	linkFormat := BuildLinksFormat(url)

	getParam := "?page="
	if strings.Contains(url, "?") {
		getParam = "&page="
	}
	var i int = 1
	var last int = 2
	var next int
	for {
		pageUrl := url + getParam + strconv.Itoa(i)
		stargazers, linkHeader, err := getStargazers(pageUrl, token)
		if err != nil {
			return timestamps, err
		}

		for _, star := range stargazers {
			var t time.Time
			t, err = time.Parse(time.RFC3339, star.Timestamp)
			if err != nil {
				return timestamps, fmt.Errorf("An error occured while parsing the timestamp: %v", err)
			}
			timestamp := t.Unix()
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
				return timestamps, fmt.Errorf("An error occured while parsing the header: %v, parser is %s, link header is %s", err, linkFormat, linkHeader)
			}
		}

		if i == last {
			break
		}
		i++
	}
	return timestamps, nil
}

func GetTimestampsDistributed(starCount, perPage int, url, token string) ([]int64, error) {
	// calculate the number of calls to make to Github API
	batchCount := starCount / perPage
	// don't forget to get the incomplete page
	if starCount%perPage > 0 {
		batchCount++
	}

	// Get the default number of GOMAXPROCS
	goprocnb := runtime.GOMAXPROCS(0)
	if goprocnb > batchCount {
		return GetTimestamps(batchCount, url, token)
	}
	fmt.Printf("Number of goroutines: %d\n", goprocnb)

	url = url + "?per_page=" + strconv.Itoa(perPage)

	// Create a mutex to be able to avoid race condition
	// And the two variables that will be protected by it.
	mutex := &sync.Mutex{}
	timestamps := make([]int64, starCount)
	var arrayCursor = 0

	// create error channel to send the errors
	// from the goroutines and the master
	errchan := make(chan error)
	defer close(errchan)

	// create intchan to send order from the master
	// to the goroutines
	intchan := make(chan int64)
	defer close(intchan)

	// create finishedchan so the goroutines
	// can tell the master that the job is finished
	// and the master will say to stop
	finishedchan := make(chan struct{})
	defer close(finishedchan)

	// create stopchans to send to the goroutine
	// the signal to stop it's work
	stopchans := make([]chan struct{}, goprocnb)
	for k := 0; k < goprocnb; k++ {
		stopchans[k] = make(chan struct{})
		defer close(stopchans[k])
	}

	// create a WaitGroup to wait for the goroutine
	// to end before finishing executing the program
	var wg sync.WaitGroup

	for j := 0; j < goprocnb; j++ {
		fmt.Printf("Start goroutine: %d\n", j)
		wg.Add(1)
		go func(goroutine int, stopchan chan struct{}) {
			for {
				select {
				case input := <-intchan:
					fmt.Printf("Goroutine %d got %d to get\n", goroutine, input)
					err := goroutineWork(&timestamps, mutex, input, &arrayCursor, url, token)
					if err != nil {
						errchan <- err
					} else {
						finishedchan <- stru{}
					}
				case <-stopchan:
					fmt.Printf("Goroutine %d recieved the order to stop\n", goroutine)
					wg.Done()
					fmt.Printf("Goroutine %d returning\n", goroutine)
					return
				}
			}
		}(j, stopchans[j])
	}

	var i int64 = 1
	fmt.Println("Sending initial requests")
	for j := 0; j < goprocnb; j++ {
		intchan <- i
		i++
	}
	fmt.Println("Initial requests sent")
L:
	for {
		select {
		case <-finishedchan:
			fmt.Println("A goroutine finished a job, incrementing i...")
			if int(i) > batchCount {
				fmt.Println("Job over, requesting goroutine stop...")
				for k := 0; k < goprocnb-1; k++ {
					<-finishedchan
					fmt.Printf("Resp recieved\n")
				}

				for k := 0; k < goprocnb; k++ {
					fmt.Printf("Sending request %d\n", k)
					stopchans[k] <- stru{}
					fmt.Printf("Request %d sent\n", k)
				}
				fmt.Println("Requests sent")
				break L
			}
			intchan <- i
			i++
		case err := <-errchan:
			fmt.Printf("An error occured while getting the timestamp: %v\n", err)
			for k := 0; k < goprocnb-1; k++ {
				<-finishedchan
				fmt.Printf("Resp recieved\n")
			}

			for k := 0; k < goprocnb; k++ {
				fmt.Printf("Sending request %d\n", k)
				stopchans[k] <- stru{}
				fmt.Printf("Request %d sent\n", k)
			}
			fmt.Println("Requests sent")
			break L
		}
	}
	fmt.Println("Waiting for the goroutines to end it")
	wg.Wait()
	sort.Sort(sortableTimestamps(timestamps))

	return timestamps, nil
}

type stargazer struct {
	Timestamp string `json:"starred_at"`
}

func BuildLinksFormat(url string) string {
	if strings.Contains(url, "?") {
		return "<" + url + "&page=%d>; rel=\"next\", <" + url + "&page=%d>; rel=\"last\""
	}
	return "<" + url + "?page=%d>; rel=\"next\", <" + url + "?page=%d>; rel=\"last\""
}

func goroutineWork(timestamps *[]int64, mutex *sync.Mutex, i int64, arrayCursor *int, url, token string) error {
	getParam := "?page="
	if strings.Contains(url, "?") {
		getParam = "&page="
	}

	pageUrl := url + getParam + strconv.Itoa(int(i))
	stargazers, _, err := getStargazers(pageUrl, token)
	if err != nil {
		return err
	}

	for _, star := range stargazers {
		var t time.Time
		t, err = time.Parse(time.RFC3339, star.Timestamp)
		if err != nil {
			return fmt.Errorf("An error occured while parsing the timestamp: %v", err)
		}
		timestamp := t.Unix()
		mutex.Lock()
		(*timestamps)[*arrayCursor] = timestamp
		(*arrayCursor) = (*arrayCursor) + 1
		mutex.Unlock()
	}

	return nil
}

func getStargazers(pageUrl, token string) ([]stargazer, string, error) {
	r, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return make([]stargazer, 0), "", fmt.Errorf("An error occured while creating the request: %v", err)
	}

	r.Header.Add("Accept", "application/vnd.github.v3.star+json")
	r.Header.Add("Authorization", "token "+token)

	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return make([]stargazer, 0), "", fmt.Errorf("An error occured while doing the request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]stargazer, 0), "", fmt.Errorf("An error occured while reading the body: %v", err)
	}

	stargazers := make([]stargazer, 0)
	err = json.Unmarshal(body, &stargazers)
	if err != nil {
		return make([]stargazer, 0), "", fmt.Errorf("An error occured while unmarshalling: %v", err)
	}
	return stargazers, resp.Header.Get("Link"), nil
}

type sortableTimestamps []int64

func (s sortableTimestamps) Len() int           { return len(s) }
func (s sortableTimestamps) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortableTimestamps) Less(i, j int) bool { return s[i] < s[j] }

type stru struct{}
