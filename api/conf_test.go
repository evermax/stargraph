package api

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/mq"
)

func TestNewConf(t *testing.T) {
	q := &msgq{}
	NewConf(storedb{}, q, "add", "update")
	if q.addQueue != "add" {
		t.Fatalf("addQueue should be \"add\", instead is %s", q.addQueue)
	}
	if q.updateQueue != "update" {
		t.Fatalf("updateQueue should be \"update\", instead is %s", q.updateQueue)
	}
}

func TestTriggerAddJob(t *testing.T) {
	var aqn = "add"
	var uqn = "update"

	q := &msgq{}
	var conf = Conf{
		MessageQueue: q,
		AddQueue:     aqn,
		UpdateQueue:  uqn,
	}

	err := conf.MessageQueue.DeclareQueue(aqn)
	if err != nil {
		t.Fatalf("An error occured while Declaring a queue: %v", err)
	}

	var count = rand.Intn(10)
	for i := 0; i < count; i++ {
		conf.TriggerAddJob(github.RepoInfo{}, "test")
	}

	if q.addJobTriggered != count {
		t.Fatalf("The number of triggered job should be: %d, but is %d", count, q.addJobTriggered)
	}
}

func TestTriggerUpdateJob(t *testing.T) {
	var aqn = "add"
	var uqn = "update"

	q := &msgq{}
	var conf = Conf{
		MessageQueue: q,
		AddQueue:     aqn,
		UpdateQueue:  uqn,
	}

	err := conf.MessageQueue.DeclareQueue(uqn)
	if err != nil {
		t.Fatalf("An error occured while Declaring a queue: %v", err)
	}

	var count = rand.Intn(10)
	for i := 0; i < count; i++ {
		conf.TriggerUpdateJob(github.RepoInfo{}, "test")
	}

	if q.updateJobTriggered != count {
		t.Fatalf("The number of triggered job should be: %d, but is %d", count, q.updateJobTriggered)
	}
}

type msgq struct {
	addQueue           string
	updateQueue        string
	addJobTriggered    int
	updateJobTriggered int
	token              string
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
	return nil
}
