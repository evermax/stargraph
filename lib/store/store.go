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

// ID struct
type ID interface {
}

// Store interface
type Store interface {
	AddRepo(github.RepoInfo) (ID, error)
	GetRepo(string) (github.RepoInfo, ID, error)
	PutRepo(github.RepoInfo, ID) error
	ClaimWork(github.RepoInfo, ID) error
}
