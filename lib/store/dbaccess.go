package store

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"

	"github.com/evermax/stargraph/github"
)

const (
	kind     = "RepoInfo"
	stringID = "default_repoinfo"
)

var (
	// ErrAlreadyExist is the error return if the Github repository is requested to be created but already exist.
	ErrAlreadyExist = fmt.Errorf("Repository already stored in database")
	// ErrAlreadyWorkedOn express the fact that the repository is already claim to work on.
	ErrAlreadyWorkedOn = fmt.Errorf("Repository already claimed in database")
)

// Store is a simple struct that hold the context to access the datastore
// and has several methods to add, get and modify informations stored about Github repositories.
type Store struct {
	Context context.Context
}

// NewStore create a new Store with a default context.
// It actually use context.Background().
func NewStore() Store {
	return Store{
		Context: context.Background(),
	}
}

// GetRepo will fetch the data about the repo from the database.
// If the repo exist, return the RepoInfo populated
// If the repo doesn't exist in the database, return Repo with Exist = false
// Return an error only if something unexepected happened.
func (store Store) GetRepo(repo string) (github.RepoInfo, *datastore.Key, error) {
	var repoInfo github.RepoInfo
	key := new(datastore.Key)
	q := datastore.NewQuery(kind).Ancestor(repoInfoKey(store.Context)).Filter("Name =", repo)
	for t := q.Run(store.Context); ; {
		var info github.RepoInfo
		var err error
		key, err = t.Next(&info)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return repoInfo, nil, err
		}
		repoInfo = info
		repoInfo.SetExist(true)
	}
	return repoInfo, key, nil
}

// AddRepo add a Github repository entry into the database.
// If the repository exist, return AlreadyExistError.
// Return an eventual error from the communication with the database.
func (store Store) AddRepo(repoInfo github.RepoInfo) (*datastore.Key, error) {
	info, _, err := store.GetRepo(repoInfo.Name)
	if err != nil {
		return nil, err
	}
	if info.Exist() {
		return nil, ErrAlreadyExist
	}
	return datastore.Put(store.Context, repoInfoKey(store.Context), &repoInfo)
}

// PutRepo will put the informations about the Github repository in the database
func (store Store) PutRepo(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	return datastore.Put(store.Context, key, &repoInfo)
}

// ClaimWork set the WorkedOn flag of the repoInfo to true and persist it.
func (store Store) ClaimWork(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	info, _, err := store.GetRepo(repoInfo.Name)
	if err != nil {
		return nil, err
	}
	if info.WorkedOn {
		return nil, ErrAlreadyWorkedOn
	}
	repoInfo.WorkedOn = true
	return datastore.Put(store.Context, key, &repoInfo)
}

// repoInfoKey returns the key used for all repoInfo entries.
func repoInfoKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, kind, stringID, 0, nil)
}
