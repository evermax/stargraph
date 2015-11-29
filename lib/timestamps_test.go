package lib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetTimestampsSimple(t *testing.T) {
	filePath := "testdata/simple_stars.json"
	expectedTimestamps := []int64{1446285600, 1446289200, 1446292800}
	body, err := ioutil.ReadFile(filePath)
	batch := 1
	if err != nil {
		t.Fatalf("An error occured while reading the file %s: %v\n", filePath, err)
	}

	serverUrl := ""
	handler := func(w http.ResponseWriter, r *http.Request) {
		linkFormat := BuildLinksFormat(serverUrl)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, 1, 1))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	serverUrl = server.URL + "?per_page=" + strconv.Itoa(batch)

	timestamps, err := GetTimestamps(1, server.URL, "token")
	if err != nil {
		t.Fatalf("An error occured while requesting the timestamps: %v\n", err)
	}
	if len(timestamps) != len(expectedTimestamps) {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same size\n", expectedTimestamps, timestamps)
	}
	equals := true
	for i, v := range timestamps {
		if expectedTimestamps[i] != v {
			equals = false
			break
		}
	}
	if !equals {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same values\n", expectedTimestamps, timestamps)
	}
}

func TestGetTimestampsMulPages(t *testing.T) {
	filePathFormat := "testdata/mul_stars_%d.json"
	expectedTimestamps := []int64{1446285600, 1446289200, 1446292800}
	maxPage := 3
	batch := 3
	serverUrl := ""
	handler := func(w http.ResponseWriter, r *http.Request) {
		page := 1
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
		linkFormat := BuildLinksFormat(serverUrl)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, page, maxPage))
		w.Write(body)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	serverUrl = server.URL + "?per_page=" + strconv.Itoa(batch)
	timestamps, err := GetTimestamps(batch, server.URL, "token")
	if err != nil {
		t.Fatalf("An error occured while requesting the timestamps: %v\n", err)
	}
	if len(timestamps) != len(expectedTimestamps) {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same size\n", expectedTimestamps, timestamps)
	}
	equals := true
	for i, v := range timestamps {
		if expectedTimestamps[i] != v {
			equals = false
			break
		}
	}
	if !equals {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same values\n", expectedTimestamps, timestamps)
	}
}

func TestGetTimestampsDistributed(t *testing.T) {
	filePathFormat := "testdata/distributed_stars_%d.json"
	expectedTimestamps := 16
	maxPage := 5
	batch := 5

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
		linkFormat := BuildLinksFormat(serverUrl)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, page, maxPage))
		w.Write(body)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	serverUrl = server.URL + "?per_page=" + strconv.Itoa(batch)
	timestamps, err := GetTimestampsDistributed(expectedTimestamps, batch, server.URL, "token")

	if err != nil {
		t.Fatalf("An error occured in GetTimestampsDistributed %v\n", err)
	}

	timestampCount := len(timestamps)
	if timestampCount != expectedTimestamps {
		t.Fatalf("Expected %d timestamps and got %d", expectedTimestamps, timestampCount)
	}
}

func TestGetStargazers(t *testing.T) {
	filePath := "testdata/simple_stars.json"
	expectedTimestamps := []stargazer{{"2015-10-31T10:00:00Z"}, {"2015-10-31T11:00:00Z"}, {"2015-10-31T12:00:00Z"}}
	body, err := ioutil.ReadFile(filePath)
	batch := 1
	if err != nil {
		t.Fatalf("An error occured while reading the file %s: %v\n", filePath, err)
	}

	serverUrl := ""
	handler := func(w http.ResponseWriter, r *http.Request) {
		linkFormat := BuildLinksFormat(serverUrl)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, 1, 1))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	serverUrl = server.URL + "?per_page=" + strconv.Itoa(batch)

	timestamps, _, err := getStargazers(server.URL, "token")
	if err != nil {
		t.Fatalf("An error occured while requesting the timestamps: %v\n", err)
	}
	if len(timestamps) != len(expectedTimestamps) {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same size\n", expectedTimestamps, timestamps)
	}
	equals := true
	for i, v := range timestamps {
		if expectedTimestamps[i] != v {
			equals = false
			break
		}
	}
	if !equals {
		t.Fatalf("The expected timestamps %v and the actual ones %v"+
			" don't have the same values\n", expectedTimestamps, timestamps)
	}
}
