package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/store"
)

func TestApiHandlerNoRepo(t *testing.T) {
	conf := Conf{}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestApiHandlerNoToken(t *testing.T) {
	conf := Conf{}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	resp, err := http.Get(server.URL + "?repo=evermax/stargraph")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestApiHandlerErrorOnDB(t *testing.T) {
	conf := Conf{
		Database: storedb{getRepoFail: true, exist: true},
	}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	req, err := http.NewRequest("GET", server.URL+"?repo=evermax/stargraph", nil)
	req.Header.Add(AuthorizationHeader, "Bearer test")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("An error occured while doing the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestApiHandlerExistTriggerUpdate(t *testing.T) {
	q := &msgq{}
	q.DeclareQueue("update")
	conf := Conf{
		UpdateQueue:  "update",
		Database:     storedb{exist: true},
		MessageQueue: q,
	}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	req, err := http.NewRequest("GET", server.URL+"?repo=evermax/stargraph", nil)
	req.Header.Add(AuthorizationHeader, "Bearer test")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("An error occured while doing the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusOK)
	}

	if q.updateJobTriggered != 1 {
		t.Fatalf("Should have triggered one job, instead triggered %d", q.updateJobTriggered)
	}
}

func TestApiHandlerErrorOnTriggering(t *testing.T) {
	q := &msgq{}
	q.DeclareQueue("update")
	conf := Conf{
		UpdateQueue:  "updatez",
		Database:     storedb{exist: true},
		MessageQueue: q,
	}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	req, err := http.NewRequest("GET", server.URL+"?repo=evermax/stargraph", nil)
	req.Header.Add(AuthorizationHeader, "Bearer test")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("An error occured while doing the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestApiHandlerErrorOnGithub(t *testing.T) {
	q := &msgq{}
	q.DeclareQueue("add")
	conf := Conf{
		UpdateQueue:  "add",
		Database:     storedb{},
		MessageQueue: q,
	}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	req, err := http.NewRequest("GET", server.URL+"?repo=evermax/stargraph", nil)
	req.Header.Add(AuthorizationHeader, "Bearer test")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("An error occured while doing the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusInternalServerError)
	}
}

type storedb struct {
	addRepoFail   bool
	getRepoFail   bool
	putRepoFail   bool
	claimWorkFail bool
	exist         bool
}

func (db storedb) AddRepo(repo github.RepoInfo) (store.ID, error) {
	if db.addRepoFail {
		return iD{}, fmt.Errorf("Random Error")
	}
	return iD{}, nil
}

func (db storedb) GetRepo(repo string) (github.RepoInfo, store.ID, error) {
	if db.getRepoFail {
		return github.RepoInfo{}, iD{}, fmt.Errorf("Random Error")
	}
	repoInfo := github.RepoInfo{}
	(&repoInfo).SetExist(db.exist)
	return repoInfo, iD{}, nil
}

func (db storedb) PutRepo(repo github.RepoInfo, id store.ID) error {
	if db.putRepoFail {
		return fmt.Errorf("Random Error")
	}
	return nil
}

func (db storedb) ClaimWork(repo github.RepoInfo, id store.ID) error {
	if db.claimWorkFail {
		return fmt.Errorf("Random Error")
	}
	return nil
}

type iD struct{}

func (id iD) Test() string {
	return ""
}
