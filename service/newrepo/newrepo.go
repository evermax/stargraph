// Package newrepo contains the creator service that should create a new entry in the database for a github repo
// that is not already in there.
package newrepo

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/evermax/stargraph/api"
	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/store"
	"github.com/evermax/stargraph/service"
	"github.com/streadway/amqp"
)

// Creator is just a wrapper of string
type Creator string

// NewCreator will create a new creator
func NewCreator() Creator {
	// connect to the AMQP server
	return Creator(service.CreatorName)
}

// Type will return the string "creator" to implement the interface.
func (c Creator) Type() string {
	return string(c)
}

// Run create a connection to the AMQP server and listen to incoming requests
// To create Github repository graphs.
func (c Creator) Run(jobQueue chan service.Job, amqpURL, addQueueN string) error {

	// Dial connection to the AMQP server
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("Failed to connect to AMQP with url %s: %v", amqpURL, err)
	}
	defer conn.Close()

	// Open a channel of communication through the connection
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare a queue
	q, err := ch.QueueDeclare(
		addQueueN, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("Failed to set QoS: %v", err)
	}

	// Register as a consumer of the queue and
	// return a channel that brings the incoming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		return fmt.Errorf("Failed to register as consumer: %v", err)
	}

	// Declare the datastore access
	stre := store.NewStore()

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			err := creatorWork(stre, jobQueue, d.Body)
			if err == store.ErrAlreadyExist {
				d.Ack(false)
				log.Printf("WARN: Asked to recreate %s, aborting", d.Body)
				continue
			}
			if err != nil {
				// log it
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
			log.Printf("Done")
		}
		forever <- true
	}()
	<-forever
	return nil
}

func creatorWork(stre store.Store, jobQueue chan service.Job, body []byte) error {
	apiJob, err := api.Unmarshal(body)
	if err != nil {
		return fmt.Errorf("Umarshalling error with %s: %v", body, err)
	}

	repoInfo, _, err := stre.GetRepo(apiJob.RepoInfo.Name)
	if err != nil {
		return fmt.Errorf("Access to store error with %s: %v", body, err)
	}

	// Create the repository on the datastore, claim the work
	repoInfo.WorkedOn = true
	key, err := stre.AddRepo(repoInfo)
	if err != nil {
		return fmt.Errorf("Adding to store error with %s: %v", body, err)
	}

	timestamps, err := GetAllTimestamps(jobQueue, 100, apiJob.Token, repoInfo)
	if err != nil {
		return fmt.Errorf("Error with %s: %v", body, err)
	}

	lastStar := timestamps[len(timestamps)-1]

	repoInfo.Timestamps = timestamps

	repoInfo.WorkedOn = false
	repoInfo.LastUpdate = time.Now().Format(time.RFC3339)
	repoInfo.LastStarDate = time.Unix(lastStar, 0).Format(time.RFC3339)
	key, err = stre.PutRepo(repoInfo, key)
	if err != nil {
		return fmt.Errorf("Put to store error with %s: %v", body, err)
	}

	// TODO: Think if this could be done on the fly first
	// TODO: lib.CanvasJS(timestamps, repoInfo, buffer)
	// Then send the buffer to Google Storage
	// service.Objects.Insert(*bucketName, object).Media(file).Do()
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
