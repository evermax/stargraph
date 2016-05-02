package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evermax/stargraph/github"
)

// WorkerPool is a simple type wrapper for chan chan Job
// So it is easier to pass/expect it.
type WorkerPool chan chan Job

// Job structure gets passed through the Job chan
// to tell one of the go routine of a worker what to do
type Job struct {
	Num               int
	ApiURL            string
	ApiToken          string
	ErrorChannel      chan error
	TimestampsChannel chan []int64
}

// Worker is the structure that will be passed to the worker pool
// It has two methods public Start and Stop that are here to start and stop the worker.
// It can be use both to get and update the stars of a Github repository
// as it only return for the stars for the given ApiURL (which contains the number of stars and the page).
// TODO should it be a interface?
type Worker struct {
	WorkerPool   WorkerPool
	JobChannel   chan Job
	workerNumber int
	quit         chan bool
}

// NewWorker create a new worker linked to the provided workerPool.
// The Worker must be manually started for it to join the worker pool.
// The number provided is only there for debugging purposes.
func NewWorker(workerPool WorkerPool, number int) Worker {
	return Worker{
		WorkerPool:   workerPool,
		JobChannel:   make(chan Job),
		quit:         make(chan bool),
		workerNumber: number,
	}
}

// Start method will start the worker.
// The worker join the WorkerPool by providing its JobChannel.
// It then listen to the JobChannel for a new job to work on.
func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				timestamps, err := job.work()
				if err != nil {
					job.ErrorChannel <- err
				} else {
					job.TimestampsChannel <- timestamps
				}
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop method will send a message to the worker to make it stop.
// It is interesting put this Stop method in a defer function.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (job Job) work() ([]int64, error) {
	getParam := "?page="
	if strings.Contains(job.ApiURL, "?") {
		getParam = "&page="
	}
	pageURL := job.ApiURL + getParam + strconv.Itoa(int(job.Num))

	stargazers, _, err := github.GetStargazers(pageURL, job.ApiToken)
	if err != nil {
		return make([]int64, 0), err
	}

	var timestamps []int64
	for _, star := range stargazers {
		timestamp, err := star.GetTimestamp()
		if err != nil {
			return make([]int64, 0), fmt.Errorf("An error occured while parsing the timestamp: %v", err)
		}

		timestamps = append(timestamps, timestamp)
	}

	return timestamps, nil
}
