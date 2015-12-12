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

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func AddRepoWork(jobQueue chan service.Job, amqpURL, addQueueN string) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("Failed to connect to AMQP with url %s: %v", amqpURL, err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

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
		return fmt.Errorf("Failed to register a consumer: %v", err)
	}

	// Declare the datastore access
	stre := store.NewStore()

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			err := serviceWork(stre, jobQueue, d.Body)
			if err == store.AlreadyExistError {
				d.Ack(false)
				log.Printf("Asked to recreate %s, aborting", d.Body)
				continue
			}
			if err != nil {
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

func serviceWork(stre store.Store, jobQueue chan service.Job, body []byte) error {
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

func GetAllTimestamps(jobQueue chan service.Job, perPage int, token string, repoInfo github.IRepoInfo) ([]int64, error) {
	// calculate the number of calls to make to Github API
	batchCount := repoInfo.GetCount() / perPage
	// don't forget to get the incomplete page
	if repoInfo.GetCount()%perPage > 0 {
		batchCount++
	}

	url := repoInfo.URL() + "?per_page=" + strconv.Itoa(perPage)

	// create the timestamp array that will be used to
	// agregate all the timestamps that the main routine gets
	// from the workers
	timestamps := make([]int64, 0)
	stampsChan := make(chan []int64, 8)
	defer close(stampsChan)

	// create error channel to send the errors
	// from the goroutines and the master
	errchan := make(chan error)
	defer close(errchan)

	var j int = 0
	var err error
	for i := 0; i < batchCount; i++ {
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
			if j >= batchCount {
				break L
			}

		case stamps := <-stampsChan:
			timestamps = append(timestamps, stamps...)
			j++
			if j >= batchCount {
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
