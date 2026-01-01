package syncqueue

import (
	"context"
	"sync"
	"time"
)

var GlobalJobQueue *Queue

// Job represents a generic task with retries and timeout
type Job struct {
	Callback func(ctx context.Context) error
	Retries  int
	Delay    time.Duration
	Timeout  time.Duration
	Done     func(err error)
}

// Queue manages a dynamic pool of jobs
type Queue struct {
	jobs       chan Job
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	maxWorkers int
	active     int
	mu         sync.Mutex
}

// New creates a new dynamic job queue
func New(maxWorkers int, bufferSize int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		jobs:       make(chan Job, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
		maxWorkers: maxWorkers,
	}

	// start with 1 worker
	q.spawnWorker()
	return q
}

// Add enqueues a new job
func (q *Queue) Add(jobFunc func(ctx context.Context) error, retries int, delay, timeout time.Duration) {
	q.AddWithDone(jobFunc, retries, delay, timeout, nil)
}

// AddWithDone enqueues a job with an optional done callback
func (q *Queue) AddWithDone(jobFunc func(ctx context.Context) error, retries int, delay, timeout time.Duration, done func(err error)) {
	select {
	case <-q.ctx.Done():
		if done != nil {
			done(context.Canceled)
		}
		return
	default:
	}

	q.wg.Add(1)
	job := Job{
		Callback: jobFunc,
		Retries:  retries,
		Delay:    delay,
		Timeout:  timeout,
		Done:     done,
	}

	q.jobs <- job

	q.mu.Lock()
	if q.active < q.maxWorkers && len(q.jobs) > q.active {
		q.spawnWorker()
	}
	q.mu.Unlock()
}

// Wait waits for all queued jobs to finish
func (q *Queue) Wait() {
	q.wg.Wait()
	q.Shutdown()
}

// Shutdown closes the queue and cancels all workers
func (q *Queue) Shutdown() {
	q.cancel()
	close(q.jobs)
}

// spawnWorker starts a new worker
func (q *Queue) spawnWorker() {
	q.active++
	go func() {
		defer func() {
			q.mu.Lock()
			q.active--
			q.mu.Unlock()
		}()

		for job := range q.jobs {
			var lastErr error
			for attempt := 0; attempt <= job.Retries; attempt++ {
				ctx := q.ctx
				var cancel context.CancelFunc
				if job.Timeout > 0 {
					ctx, cancel = context.WithTimeout(ctx, job.Timeout)
				}

				err := job.Callback(ctx)
				if cancel != nil {
					cancel()
				}

				if err == nil {
					lastErr = nil
					break
				}

				lastErr = err
				if attempt < job.Retries {
					select {
					case <-time.After(job.Delay):
					case <-q.ctx.Done():
						lastErr = context.Canceled
						break
					}
				}
			}

			if job.Done != nil {
				job.Done(lastErr)
			}
			q.wg.Done()
		}
	}()
}
