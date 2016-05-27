package store

import (
	"fmt"

	"github.com/evermax/stargraph/github"
)

var (
	// ErrAlreadyExist is the error return if the Github repository is requested to be created but already exist.
	ErrAlreadyExist = fmt.Errorf("Repository already stored in database")
	// ErrAlreadyWorkedOn express the fact that the repository is already claim to work on.
	ErrAlreadyWorkedOn = fmt.Errorf("Repository already claimed in database")
)

// ID interface holds whatever system of id the underlying
// database uses.
// TODO complete this to something that actually makes sense
type ID interface {
	Test() string
}

// Store interface allows this service to rely on different types of databases.
// The reason for it is scalability and of course testability.
// At first, you might what to deploy everthing on a simple computer with a postgres instance.
// Then, the easiest might be to migrate to AWS or Google Cloud and have different type of database.
// In any case, the rest of the project only needs a simple interface for that, the Store interface.
// TODO add a RemoveRepo method.
type Store interface {
	AddRepo(github.RepoInfo) (ID, error)
	GetRepo(string) (github.RepoInfo, ID, error)
	PutRepo(github.RepoInfo, ID) error
	ClaimWork(github.RepoInfo, ID) error
}
