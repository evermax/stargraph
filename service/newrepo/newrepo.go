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

func AddRepoWork(jobQueue chan service.Job, amqpURL, addQueueN string) {
	conn, err := amqp.Dial(amqpURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		addQueueN, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			apiJob, err := api.Unmarshal([]byte(d.Body))
			if err != nil {
				d.Nack(false, true)
				log.Printf("Error while unmarshalling the Job: %v", err)
				continue
			}

			stre := store.NewStore()
			repoInfo, _, err := stre.GetRepo(apiJob.RepoInfo.Name)
			if err != nil {
				d.Nack(false, true)
				log.Printf("Error while getting the repo: %v", err)
				continue
			}

			// Create the repository on the datastore, claim the work
			repoInfo.WorkedOn = true
			key, err := stre.AddRepo(repoInfo)
			if err == store.AlreadyExistError {
				d.Ack(false)
				log.Printf("Asked to recreate %v, aborting", repoInfo)
				continue
			}
			if err != nil {
				d.Nack(false, true)
				log.Printf("Error while creating the repo: %v\t%v", repoInfo, err)
				continue
			}

			timestamps, err := GetAllTimestamps(jobQueue, 100, apiJob.Token, repoInfo)
			if err != nil {
				d.Nack(false, true)
				log.Printf("Error while getting the timestamps for %v\t%v", repoInfo, err)
			}

			lastStar := timestamps[len(timestamps)-1]

			repoInfo.Timestamps = timestamps

			repoInfo.WorkedOn = false
			repoInfo.LastUpdate = time.Now().Format(time.RFC3339)
			repoInfo.LastStarDate = time.Unix(lastStar, 0).Format(time.RFC3339)
			key, err = stre.PutRepo(repoInfo, key)
			if err != nil {
				d.Nack(false, true)
				log.Printf("Error while persisting the repo: %v\t%v", repoInfo, err)
			}

			// TODO: lib.CanvasJS(timestamps, repoInfo, buffer)
			// Then send the buffer to Google Storage
			// service.Objects.Insert(*bucketName, object).Media(file).Do()
			// https://cloud.google.com/storage/docs/json_api/v1/json-api-go-samples

			d.Ack(false)
			log.Printf("Done")
		}
	}()
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
