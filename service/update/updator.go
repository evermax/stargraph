package update

import (
	"github.com/evermax/stargraph/service"
)

// Updator is just a wrapper of string
type Updator string

// NewUpdator will create a new updator
func NewUpdator() Updator {
	return Updator(service.UpdatorName)
}

// Type will return the string "updator" to implement the interface.
func (u Updator) Type() string {
	return string(u)
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
