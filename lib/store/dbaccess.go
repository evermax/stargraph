package store

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"

	"github.com/evermax/stargraph/github"
)

const (
	Kind     = "RepoInfo"
	StringID = "default_repoinfo"
)

var (
	AlreadyExistError = fmt.Errorf("Repository already stored in database")
)

type Store struct {
	Context context.Context
}

func NewStore() Store {
	return Store{
		Context: context.Background(),
	}
}

// If the repo exist, return the RepoInfo populated
// If the repo doesn't exist in the database, return Repo with Exist = false
// Return an error only if something unexepected happened.
func (store Store) GetRepo(repo string) (github.RepoInfo, *datastore.Key, error) {
	var repoInfo github.RepoInfo
	key := new(datastore.Key)
	q := datastore.NewQuery(Kind).Ancestor(repoInfoKey(store.Context)).Filter("Name =", repo)
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

// Add a repository
func (store Store) AddRepo(repoInfo github.RepoInfo) (*datastore.Key, error) {
	info, _, err := store.GetRepo(repoInfo.Name)
	if err != nil {
		return nil, err
	}
	if info.Exist() {
		return nil, AlreadyExistError
	}
	return datastore.Put(store.Context, repoInfoKey(store.Context), &repoInfo)
}

func (store Store) PutRepo(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	return datastore.Put(store.Context, key, &repoInfo)
}

func (store Store) ClaimWork(repoInfo github.RepoInfo, key *datastore.Key) (*datastore.Key, error) {
	repoInfo.WorkedOn = true
	return datastore.Put(store.Context, key, &repoInfo)
}

// repoInfoKey returns the key used for all repoInfo entries.
func repoInfoKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, Kind, StringID, 0, nil)
}
