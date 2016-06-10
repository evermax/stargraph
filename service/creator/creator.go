// Package creator contains the creator service that should create a new entry in the database for a github repo
// that is not already in there.
package creator

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/evermax/stargraph/api"
	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/mq"
	"github.com/evermax/stargraph/lib/store"
	"github.com/evermax/stargraph/service"
)

// Creator contains the database to use, the type of service (creator)
// and the job queue to send job to workers. It implements the service.SWorker interface.
type Creator struct {
	t         string
	db        store.Store
	messageQ  mq.MessageQueue
	jobQueue  chan service.Job
	queueName string
}

// NewCreator creates a new creator
func NewCreator(db store.Store, queue mq.MessageQueue) Creator {
	return Creator{
		t:        service.CreatorName,
		db:       db,
		messageQ: queue,
	}
}

// Type will return the string "creator" to implement the interface.
func (c Creator) Type() string {
	return c.t
}

func (c Creator) JobQueue() chan service.Job {
	return c.jobQueue
}

// Run create a connection to the AMQP server and listen to incoming requests
// To create Github repository graphs.
func (c Creator) Run() error {
	return c.messageQ.Consume(c.queueName, c.receiveMessage)
}

func (c Creator) receiveMessage(d mq.Delivery, forever chan bool) {
	body := d.Body()
	log.Printf("Received a message: %s", body)

	err := c.creatorWork(d.Body())
	if err == store.ErrAlreadyExist {
		d.Ack(false)
		log.Printf("WARN: Asked to recreate %s, aborting", d.Body)
		return
	}
	if err != nil {
		// log it
		d.Nack(false, true)
		return
	}

	d.Ack(false)
	log.Printf("Done")
}

func (c Creator) creatorWork(body []byte) error {
	apiJob, err := api.Unmarshal(body)
	if err != nil {
		return fmt.Errorf("Umarshalling error with %s: %v", body, err)
	}

	repoInfo := github.RepoInfo{
		ID:           apiJob.RepoInfo.ID,
		Name:         apiJob.RepoInfo.Name,
		Count:        apiJob.RepoInfo.Count,
		CreationDate: apiJob.RepoInfo.CreationDate,
		WorkedOn:     true,
	}

	// Create the repository on the store, claim the work
	key, err := c.db.AddRepo(repoInfo)
	if err != nil {
		return err
	}

	timestamps, err := GetAllTimestamps(c.jobQueue, 100, apiJob.Token, repoInfo)
	if err != nil {
		return fmt.Errorf("Error with %s: %v", body, err)
	}

	lastStar := timestamps[len(timestamps)-1]

	repoInfo.Timestamps = timestamps

	repoInfo.WorkedOn = false
	repoInfo.LastUpdate = time.Now().Format(time.RFC3339)
	repoInfo.LastStarDate = time.Unix(lastStar, 0).Format(time.RFC3339)
	err = c.db.PutRepo(repoInfo, key)
	if err != nil {
		return fmt.Errorf("Put to store error with %s: %v", body, err)
	}

	// TODO: Think if this could be done on the fly first
	// TODO: lib.CanvasJS(timestamps, repoInfo, buffer)
	// Then send the buffer to the database
	// TODO: wrap that into the dbaccess file service.Objects.Insert(*bucketName, object).Media(file).Do()
	// https://cloud.google.com/storage/docs/json_api/v1/json-api-go-samples
	return nil
}

// GetAllTimestamps will get the timestamps for all the stars of the passed repository.
// It will use the perPage number and the Github API token to make a number of queries the the Github API.
// The jobQueue is used to have a pool of workers that will make one API call at a time each.
// The service itself would typically share ressources with several other services.
func GetAllTimestamps(jobQueue chan service.Job, perPage int, token string, repoInfo github.IRepoInfo) ([]int64, error) {
	// calculate the number of calls to make to Github API
	numberOfAPICall := repoInfo.StarCount() / perPage
	// don't forget to add the possible incomplete page
	if repoInfo.StarCount()%perPage > 0 {
		numberOfAPICall++
	}

	url := repoInfo.URL() + "?per_page=" + strconv.Itoa(perPage)

	// create the timestamp array that will be used to
	// agregate all the timestamps that the main routine gets
	// from the workers
	var timestamps []int64
	stampsChan := make(chan []int64, 8)
	defer close(stampsChan)

	// create error channel to send the errors
	// from the goroutines and the master
	errchan := make(chan error)
	defer close(errchan)

	var j int
	var err error
	// Put jobs to make API calls in the job queue
	for i := 0; i < numberOfAPICall; i++ {
		jobQueue <- service.Job{
			Num:               i + 1,
			ApiURL:            url,
			ApiToken:          token,
			ErrorChannel:      errchan,
			TimestampsChannel: stampsChan,
		}
	}
L:
	for {
		select {
		case err = <-errchan:
			j++
			if j >= numberOfAPICall {
				break L
			}

		case stamps := <-stampsChan:
			timestamps = append(timestamps, stamps...)
			j++
			if j >= numberOfAPICall {
				break L
			}
		}
	}
	sort.Sort(sortableTimestamps(timestamps))

	return timestamps, err
}

type sortableTimestamps []int64

func (s sortableTimestamps) Len() int           { return len(s) }
func (s sortableTimestamps) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortableTimestamps) Less(i, j int) bool { return s[i] < s[j] }
