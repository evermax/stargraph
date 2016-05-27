// Package gestore contains the implementation of the store.Store interface
// for the Google Datastore.
package gestore

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/store"
)

const (
	kind     = "RepoInfo"
	stringID = "default_repoinfo"
)

// Datastore is a simple struct that hold the context to access the datastore
// and has several methods to add, get and modify informations stored about Github repositories.
type Datastore struct {
	Context context.Context
}

// NewDatastore create a new Datastore with a default context.
// It actually use context.Background().
func NewDatastore() Datastore {
	return Datastore{
		Context: context.Background(),
	}
}

// GetRepo will fetch the data about the repo from the database.
// If the repo exist, return the RepoInfo populated
// If the repo doesn't exist in the database, return Repo with Exist = false
// Return an error only if something unexepected happened.
func (db Datastore) GetRepo(repo string) (github.RepoInfo, *datastore.Key, error) {
	var repoInfo github.RepoInfo
	key := new(datastore.Key)
	q := datastore.NewQuery(kind).Ancestor(repoInfoKey(db.Context)).Filter("Name =", repo)
	for t := q.Run(db.Context); ; {
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
func (db Datastore) AddRepo(repoInfo github.RepoInfo) (*datastore.Key, error) {
	info, _, err := db.GetRepo(repoInfo.Name)
	if err != nil {
		return nil, err
	}
	if info.Exist() {
		return nil, store.ErrAlreadyExist
	}
	return datastore.Put(db.Context, repoInfoKey(db.Context), &repoInfo)
}

// PutRepo will put the informations about the Github repository in the database
func (db Datastore) PutRepo(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	return datastore.Put(db.Context, key, &repoInfo)
}

// ClaimWork set the WorkedOn flag of the repoInfo to true and persist it.
func (db Datastore) ClaimWork(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	info, _, err := db.GetRepo(repoInfo.Name)
	if err != nil {
		return nil, err
	}
	if info.WorkedOn {
		return nil, store.ErrAlreadyWorkedOn
	}
	repoInfo.WorkedOn = true
	return datastore.Put(db.Context, key, &repoInfo)
}

// repoInfoKey returns the key used for all repoInfo entries.
func repoInfoKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, kind, stringID, 0, nil)
}
