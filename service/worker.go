package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/evermax/stargraph/github"
)

// A simple type wrapper for chan chan Job
// So it is easier to pass/expect it.
type WorkerPool chan chan Job

// The job structure gets passed through the Job chan
// to tell one of the go routine of a worker what to do
type Job struct {
	Num               int
	ApiURL            string
	ApiToken          string
	ErrorChannel      chan error
	TimestampsChannel chan []int64
}

type Worker struct {
	WorkerPool   WorkerPool
	JobChannel   chan Job
	workerNumber int
	quit         chan bool
}

func NewWorker(workerPool WorkerPool, number int) Worker {
	return Worker{
		WorkerPool:   workerPool,
		JobChannel:   make(chan Job),
		quit:         make(chan bool),
		workerNumber: number,
	}
}

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
	pageUrl := job.ApiURL + getParam + strconv.Itoa(int(job.Num))

	stargazers, _, err := github.GetStargazers(pageUrl, job.ApiToken)
	if err != nil {
		return make([]int64, 0), err
	}

	timestamps := make([]int64, 0)
	for _, star := range stargazers {
		timestamp, err := star.GetTimestamp()
		if err != nil {
			return make([]int64, 0), fmt.Errorf("An error occured while parsing the timestamp: %v", err)
		}

		timestamps = append(timestamps, timestamp)
	}

	return timestamps, nil
}
