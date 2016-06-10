package creator

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/mq"
	"github.com/evermax/stargraph/lib/store"
	"github.com/evermax/stargraph/service"
)

type mockRepoInfo struct {
	count int
	url   string
}

func (info mockRepoInfo) StarCount() int {
	return info.count
}

func (info mockRepoInfo) URL() string {
	return info.url
}

func TestCreatorWorkNonJSONMessage(t *testing.T) {
	var db = storedb{}
	var d = &delvry{body: []byte("Hello, world")}
	var q = &msgq{delivery: d}
	var creator = NewCreator(db, q)
	creator.Run()

	if !d.nack {
		t.Fatalf("Nack should have been called on the delivery")
	}

	if d.ack {
		t.Fatal("Ack shouldn't have been called on the delivery")
	}
}

/*func TestCreatorWorkNonAPIJobMessage(t *testing.T) {
	var db = storedb{}
	var d = &delvry{body: []byte("{\"test\": \"test\"}")}
	var q = &msgq{delivery: d}
	var creator = NewCreator(db, q)
	creator.Run()

	if !d.nack {
		t.Fatalf("Nack should have been called on the delivery")
	}

	if d.ack {
		t.Fatal("Ack shouldn't have been called on the delivery")
	}
}

func TestCreatorWorkRepoAlreadyExist(t *testing.T) {
	var db = storedb{}
	var d = &delvry{body: []byte("{\"name\": \"evermax/stargraph\"}")}
	var q = &msgq{delivery: d}
	var creator = NewCreator(db, q)
	creator.Run()

	if !d.ack {
		t.Fatal("Ack should have been called on the delivery")
	}
	if d.nack {
		t.Fatalf("Nack shouldn't have been called on the delivery")
	}

}*/

func TestGetAllTimestamps(t *testing.T) {
	filePathFormat := "testdata/distributed_stars_%d.json"
	expectedTimestamps := 16
	maxPage := 4
	batch := 5

	// Create worker queue
	maxQueue := 4
	maxWorker := 4
	dispatch := service.NewDispatcher(maxWorker, maxQueue)
	dispatch.Run()

	var serverURL string
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
		linkFormat := github.BuildLinksFormat(serverURL)
		w.Header().Add("Link", fmt.Sprintf(linkFormat, page, maxPage))
		w.Write(body)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	serverURL = server.URL + "?per_page=" + strconv.Itoa(batch)
	repoInfo := mockRepoInfo{
		count: expectedTimestamps,
		url:   serverURL,
	}
	timestamps, err := GetAllTimestamps(dispatch.JobQueue, batch, "token", repoInfo)

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

type storedb struct {
	addRepoFail   bool
	getRepoFail   bool
	putRepoFail   bool
	claimWorkFail bool
	exist         bool
}

var ErrNoName error

func (db storedb) AddRepo(repo github.RepoInfo) (store.ID, error) {
	if repo.Name == "evermax/stargraph" {
		return iD{}, store.ErrAlreadyExist
	}
	if repo.Name == "" {
		return iD{}, ErrNoName
	}
	if repo.ID == 0 {
		return iD{}, ErrNoName
	}
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

type msgq struct {
	addQueue           string
	updateQueue        string
	addJobTriggered    int
	updateJobTriggered int
	token              string
	delivery           *delvry
}

func (q *msgq) DeclareQueue(name string) error {
	if name == "" {
		return fmt.Errorf("Name is null")
	}
	if name == "add" {
		q.addQueue = name
	}
	if name == "update" {
		q.updateQueue = name
	}
	return nil
}

func (q *msgq) Publish(name string, body []byte) error {
	if q.addQueue != name && q.updateQueue != name {
		// TODO check if indeed an non existing queue would indeed trigger error
		return fmt.Errorf("No such Q %s", name)
	}
	if name == "add" {
		q.addJobTriggered++
	}
	if name == "update" {
		q.updateJobTriggered++
	}
	return nil
}

func (q *msgq) Consume(name string, r mq.Receiver) error {
	var forever chan bool
	r(q.delivery, forever)

	return nil
}

type delvry struct {
	body []byte
	ack  bool
	nack bool
}

func (d *delvry) Body() []byte {
	return d.body
}

func (d *delvry) Ack(multiple bool) error {
	d.ack = true
	return nil
}
func (d *delvry) Nack(multiple, requeue bool) error {
	d.nack = true
	return nil
}
