package update

import (
	"github.com/evermax/stargraph/lib/mq"
	"github.com/evermax/stargraph/lib/store"
	"github.com/evermax/stargraph/service"
)

// Updator ... TODO
type Updator struct {
	t         string
	db        store.Store
	messageQ  mq.MessageQueue
	jobQueue  chan service.Job
	queueName string
}

// NewUpdator will create a new updator
func NewUpdator(db store.Store, queue mq.MessageQueue) Updator {
	return Updator{
		t:        service.UpdatorName,
		db:       db,
		messageQ: queue,
	}
}

// Run create a connection to the AMQP server and listen to incoming requests
// To update Github repository graphs.
func (u Updator) Run() {
	// Get all the timestamps from the database
	// Get the number of pages and the repo info from Github
	// Compare and go back until it is alright
	// Update the DB.
	// Set it to not worked on anymore
}

// JobQueue ... TODO
func (u Updator) JobQueue() chan service.Job {
	return u.jobQueue
}
