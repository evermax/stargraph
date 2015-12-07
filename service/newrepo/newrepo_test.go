package newrepo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/service"
)

type mockRepoInfo struct {
	count int
	url   string
}

func (info mockRepoInfo) GetCount() int {
	return info.count
}

func (info mockRepoInfo) URL() string {
	return info.url
}

func TestGetAllTimestamps(t *testing.T) {
	filePathFormat := "testdata/distributed_stars_%d.json"
	expectedTimestamps := 16
	maxPage := 4
	batch := 5

	// Create worker queue
	maxQueue := 4
	maxWorker := 4
	jobQueue := make(chan service.Job, maxQueue)
	dispatch := service.NewDispatcher(maxWorker, jobQueue)
	dispatch.Run()

	serverUrl := ""
	handler := func(w http.ResponseWriter, r *http.Request) {
		var page int
		pageString := r.FormValue("page")
		if pageString != "" {
			var err error
			page, err = strconv.Atoi(pageString)
			if err != nil {
				t.Errorf("Parsing error of %s: %v\n", pageString, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		filePath := fmt.Sprintf(filePathFormat, page)
		body, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Errorf("Reading error of %s: %v\n", filePath, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		linkFormat := github.BuildLinksFormat(serverUrl)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, page, maxPage))
		w.Write(body)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	serverUrl = server.URL + "?per_page=" + strconv.Itoa(batch)
	repoInfo := mockRepoInfo{
		count: expectedTimestamps,
		url:   serverUrl,
	}
	timestamps, err := GetAllTimestamps(jobQueue, batch, "token", repoInfo)

	if err != nil {
		dispatch.Stop()
		t.Fatalf("An error occured in GetAllTimestamps %v\n", err)
	}

	timestampCount := len(timestamps)
	if timestampCount != expectedTimestamps {
		dispatch.Stop()
		t.Fatalf("Expected %d timestamps and got %d", expectedTimestamps, timestampCount)
	}
	dispatch.Stop()
}
