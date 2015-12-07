package api

import (
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/evermax/stargraph/github"
	"github.com/streadway/amqp"
)

// A structure that hold a AMQP connection and channel
// It also contain the name of the add repo queue
// and the name of the update repo queue
type Conf struct {
	Context     context.Context
	Channel     *amqp.Channel
	Conn        *amqp.Connection
	AddQueue    string
	UpdateQueue string
}

// Start AMQP Connection, open channel of connexion
// Create 2 queues, one to send a new repo job, one to ask for existing repo updates
func NewConf(amqpURL, addQueueN, updateQueueN string) (Conf, error) {
	var conf Conf
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return conf, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return conf, err
	}
	addq, err := ch.QueueDeclare(
		addQueueN, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return conf, err
	}
	upq, err := ch.QueueDeclare(
		updateQueueN, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return conf, err
	}
	conf = Conf{
		Context:     context.Background(),
		Conn:        conn,
		Channel:     ch,
		AddQueue:    addq.Name,
		UpdateQueue: upq.Name,
	}
	return conf, err
}

// Trigger a new add job to the queue in the conf
// With the provided repoInfo and the token
func (conf Conf) TriggerAddJob(repoInfo github.RepoInfo, token string) error {
	// Create new Job from the repo info and the token
	job := NewJob(repoInfo, token)
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	// Send to job queue via AMQP
	return conf.Channel.Publish(
		"",            // exchange
		conf.AddQueue, // routing key
		false,         // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}

// Trigger a new update job to the queue in the conf
// With the provided repoInfo and the token
func (conf Conf) TriggerUpdateJob(repoInfo github.RepoInfo, token string) error {
	// Create new Job from the repo info and the token
	job := NewJob(repoInfo, token)
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	// Send to job queue via AMQP
	return conf.Channel.Publish(
		"",               // exchange
		conf.UpdateQueue, // routing key
		false,            // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}

type Job struct {
	RepoInfo github.RepoInfo
	Token    string
}

func NewJob(repoInfo github.RepoInfo, token string) Job {
	return Job{RepoInfo: repoInfo, Token: token}
}

func (j Job) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

func Unmarshal(b []byte) (Job, error) {
	var j Job
	err := json.Unmarshal(b, &j)
	return j, err
}
