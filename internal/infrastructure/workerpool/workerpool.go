package workerpool

import (
	"log"
	"log/slog"
	"strconv"
	"sync"
)

// Job represents a unit of work that can be processed by the worker pool.
// Each Job must implement the Process method.
type Job interface {
	// Process executes the job's logic.
	// It should return an error if the job fails.
	Process() error
}

// JobSubmiter contains Submit method whicj will me imlemented by WorkerPool
// It was necessary to create this for testing.
type JobSubmiter interface {
	Submit(job Job)
}

// WorkerPool manages a fixed number of workers that process
// submitted jobs concurrently.
type WorkerPool struct {
	workers int             // number of worker goroutines
	jobs    chan Job        // channel used to queue jobs
	wg      *sync.WaitGroup // wait group to track job completion
}

func NewWorkerPool(workersStr, sizeStr string, wg *sync.WaitGroup) *WorkerPool {
	workers, _ := strconv.Atoi(workersStr)
	size, _ := strconv.Atoi(sizeStr)

	return &WorkerPool{
		workers: workers,
		jobs:    make(chan Job, size),
		wg:      wg,
	}
}

// worker runs as a goroutine and continuously processes jobs
// received from the job channel until the channel is closed.
func (wp *WorkerPool) worker(i int) {
	for job := range wp.jobs {
		slog.Info("Worker processes job", "worker", i)
		err := job.Process()
		if err != nil {
			log.Println("Error while processing the job:", err)
			slog.Warn("Error while processing the job:", "error", err)
		}
		wp.wg.Done()
	}
}

// Start launches all worker goroutines.
// This method should be called before submitting jobs.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		go wp.worker(i)
	}
}

// Submit adds a job to the worker pool queue.
// It increments the WaitGroup counter before enqueuing the job.
func (wp *WorkerPool) Submit(job Job) {
	wp.wg.Add(1)
	wp.jobs <- job
}

// Shutdown closes the job channel, signaling workers
// that no more jobs will be submitted.
func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
}

// Wait blocks until all submitted jobs have finished processing.
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
