package api

import (
	"encoding/json"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/mq"
	"github.com/evermax/stargraph/lib/store"
)

// Conf holds a mq.MessageQueue, the name of the add repo queue
// and the name of the update repo queue.
type Conf struct {
	Database     store.Store
	MessageQueue mq.MessageQueue
	AddQueue     string
	UpdateQueue  string
}

// NewConf start AMQP Connection, open channel of connexion
// Create 2 queues, one to send a new repo job, one to ask for existing repo updates
func NewConf(db store.Store, messageQ mq.MessageQueue, addQueueN, updateQueueN string) (conf Conf, err error) {
	err = messageQ.DeclareQueue(addQueueN)
	if err != nil {
		return
	}
	err = messageQ.DeclareQueue(updateQueueN)
	if err != nil {
		return
	}
	conf = Conf{
		Database:     db,
		MessageQueue: messageQ,
		AddQueue:     addQueueN,
		UpdateQueue:  updateQueueN,
	}
	return
}

// TriggerAddJob triggers a new add job to the queue in the conf
// With the provided repoInfo and the token
func (conf Conf) TriggerAddJob(repoInfo github.RepoInfo, token string) error {
	// Create new Job from the repo info and the token
	job := NewJob(repoInfo, token)
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	// Send to job queue via AMQP
	return conf.MessageQueue.Publish(conf.AddQueue, body)
}

// TriggerUpdateJob trigger a new update job to the queue in the conf
// With the provided repoInfo and the token
func (conf Conf) TriggerUpdateJob(repoInfo github.RepoInfo, token string) error {
	// Create new Job from the repo info and the token
	job := NewJob(repoInfo, token)
	body, err := job.Marshal()
	if err != nil {
		return err
	}

	// Send to job queue via AMQP
	return conf.MessageQueue.Publish(conf.UpdateQueue, body)
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
