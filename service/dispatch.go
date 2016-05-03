package service

// Dispatcher structure has a pool of workers tand will dispatch incoming
// jobs from the JobQueue to one of the workers via the WorkerPool channel
type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	JobQueue   chan Job
	workers    []Worker
	maxWorkers int
}

// NewDispatcher is a wrapper to create an new dispatch with only specifying
// the maximum number of workers and the size of the queue. The best is that
// the queue is at least the size of the number of workers but it can be more 
// if you don't want an add to be blocking.
func NewDispatcher(maxWrkrs, maxQueue int) *Dispatcher {
	pool := make(chan chan Job, maxWrkrs)
	jobQueue := make(chan Job, maxQueue)
	return &Dispatcher{
		WorkerPool: pool,
		JobQueue:   jobQueue,
		maxWorkers: maxWrkrs,
	}
}

// Run will create all the workers and start them to wait for a job to arrive 
// from the JobQueue. The dispatcher will then dispach the job to an available worker
func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool, i)
		worker.Start()
		d.workers = append(d.workers, worker)
	}

	go d.dispatch()
}

// Stop is a method that need to be called before exiting the program.
// the Dispatcher needs to be stopped because there are several goroutines
// started by it: the dispatcher and the workers.
func (d *Dispatcher) Stop() {
	for _, w := range d.workers {
		w.Stop()
	}
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
}
