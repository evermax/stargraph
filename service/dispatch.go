package service

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan chan Job
	JobQueue   chan Job
	workers    []Worker
	maxWorkers int
}

func NewDispatcher(maxWrkrs int, jq chan Job) *Dispatcher {
	pool := make(chan chan Job, maxWrkrs)
	return &Dispatcher{
		WorkerPool: pool,
		JobQueue:   jq,
		maxWorkers: maxWrkrs,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool, i)
		worker.Start()
		d.workers = append(d.workers, worker)
	}

	go d.dispatch()
}

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
