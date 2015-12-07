package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/evermax/stargraph/github"
	"github.com/evermax/stargraph/lib/store"
)

const (
	AuthorizationHeader string = "Authorization"
	ContentTypeHeader   string = "Content-Type"
	JSONContentHeader   string = "application/json; charset=utf8"
	RepoParameter       string = "repo"
)

var (
	MissingRepoError  = ErrorMessage{Error: "Repo parameter missing"}
	MissingTokenError = ErrorMessage{Error: "Token header missing"}
	InternalError     = ErrorMessage{Error: "Sorry, internal server error"}
	NotFoundError     = ErrorMessage{Error: "Repo not on Github"}
)

func (conf Conf) ApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(ContentTypeHeader, JSONContentHeader)

	repo := r.FormValue("repo")
	if repo == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MissingRepoError)
		return
	}

	tokenHeader := strings.Split(r.Header.Get(AuthorizationHeader), " ")
	if len(tokenHeader) <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(MissingTokenError)
		return
	}
	token := tokenHeader[1]

	// Get the data with the token from db.
	database := store.Store{
		Context: conf.Context,
	}

	repoInfo, _, err := database.GetRepo(repo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(InternalError)
		return
	}

	// if exist
	if repoInfo.Exist() {
		if err := conf.TriggerUpdateJob(repoInfo, token); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(InternalError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(repoInfo)
		return
	}

	// if doesn't exist in db, check on github
	repoInfo, err = github.GetRepoInfo(repo, token)
	if repoInfo.Exist() {
		if err := conf.TriggerAddJob(repoInfo, token); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(InternalError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(repoInfo)
	}

	// if doesn't exist on github 404
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(NotFoundError)
}

type ErrorMessage struct {
	Error string
}
