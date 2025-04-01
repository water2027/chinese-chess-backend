package utils

import (
	"context"
	"runtime"
	"sync"
)

type Task func() error
type Job struct {
	Task  Task
}

type WorkerPool struct {
	WorkerCount int
	JobQueue    chan Job
	ErrChan     chan error
	stopCh      chan struct{}
	once        sync.Once
	wg          sync.WaitGroup
}

func NewWorkerPool() *WorkerPool {
	workerCount := max(runtime.NumCPU(), 1)
	queueSize := workerCount * 10
	return &WorkerPool{
		WorkerCount: workerCount,
		JobQueue:    make(chan Job, queueSize),
		stopCh:      make(chan struct{}),
	}
}

func (wp *WorkerPool) Process(ctx context.Context, task Task) {
	select {
	case wp.JobQueue <- Job{Task: task}:
	case <-ctx.Done():

	}
}

func (wp *WorkerPool) Start() {
	wp.once.Do(func() {
		wp.wg.Add(wp.WorkerCount)
		for range wp.WorkerCount {
			go func() {
				defer wp.wg.Done()
				for {
					select {
					case job, ok := <-wp.JobQueue:
						if !ok {
							return
						}
						var err error
						func() {
							defer func() {
								if r := recover(); r != nil {
									if e, ok := r.(error); ok {
										err = e
									}
								}
							}()
							err = job.Task()
						}()
						if err != nil {
							wp.ErrChan <- err
						}
					case <-wp.stopCh:
						return
					}
				}
			}()
		}
	})
}

func (wp *WorkerPool) Stop() {
	close(wp.stopCh)
	wp.wg.Wait()
}
