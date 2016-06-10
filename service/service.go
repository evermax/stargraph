package service

import (
	"log"
	"net/http"
	"text/template"
)

const (
	CreatorName = "creator"
	UpdatorName = "updator"
)

// SWorker interface will be use by the service runner programm
// to run either a Creator and an Updator without any prior knowledge
// of the service being one or the other
type SWorker interface {
	Type() string
	Run()
	JobQueue() chan Job
}

// Service structure
type Service struct {
	Dispatcher *Dispatcher
}

// NewService TODO desc
func NewService(workers []SWorker, amqpURL, creatorQueue, updatorQueue, address string, maxWorker, jobQSize int) {
	dispatcher := NewDispatcher(maxWorker, jobQSize)
	defer dispatcher.Stop()
	for _, worker := range workers {
		go worker.Run()
	}
	// Start HTTP server here for status check and the one for heartbeat
	service := Service{
		Dispatcher: dispatcher,
	}
	r := http.NewServeMux()
	r.HandleFunc("/", service.uiHandler)
	log.Println(http.ListenAndServe(address, r))
}

func (s Service) uiHandler(rw http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/status.html")
	if err != nil {
		log.Panic("Error occured parsing the template", err)
	}
	if err = tmpl.Execute(rw, s.Dispatcher.Status()); err != nil {
		log.Panic("Failed to write template", err)
	}
}
