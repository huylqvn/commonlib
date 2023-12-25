package async

import (
	"runtime"
	"sync"
)

var DefaultPoolSize = PoolSize(8)

func PoolSize(size int) int {
	return size * runtime.NumCPU()
}

type Worker struct {
	done             *sync.WaitGroup
	readyPool        chan chan Job
	assignedJobQueue chan Job

	quit chan bool
}

func NewWorker(readyPool chan chan Job, done *sync.WaitGroup) *Worker {
	return &Worker{
		done:             done,
		readyPool:        readyPool,
		assignedJobQueue: make(chan Job),
		quit:             make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		w.done.Add(1)
		for {
			w.readyPool <- w.assignedJobQueue
			select {
			case job := <-w.assignedJobQueue:
				job.Process()
			case <-w.quit:
				w.done.Done()
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.quit <- true
}

type Job interface {
	Process()
}

type JobQueue struct {
	internalQueue     chan Job
	readyPool         chan chan Job
	workers           []*Worker
	dispatcherStopped *sync.WaitGroup
	workersStopped    *sync.WaitGroup
	quit              chan bool
	stopped           bool
	stoppedMutex      *sync.Mutex
}

func NewJobQueue(maxWorkers int) *JobQueue {
	workersStopped := &sync.WaitGroup{}
	readyPool := make(chan chan Job, maxWorkers)
	workers := make([]*Worker, maxWorkers, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = NewWorker(readyPool, workersStopped)
	}
	return &JobQueue{
		internalQueue:     make(chan Job),
		readyPool:         readyPool,
		workers:           workers,
		dispatcherStopped: &sync.WaitGroup{},
		workersStopped:    workersStopped,
		quit:              make(chan bool),
	}
}

func (q *JobQueue) Start() {
	for i := 0; i < len(q.workers); i++ {
		q.workers[i].Start()
	}
	go q.dispatch()

	q.stoppedMutex = &sync.Mutex{}
	q.setStopped(false)
}

func (q *JobQueue) Stop() {
	q.setStopped(true)

	// Stopping queue
	q.quit <- true
	q.dispatcherStopped.Wait()

	// Stopped queue
	close(q.internalQueue)
}

func (q *JobQueue) Stopped() bool {
	return q.getStopped()
}

func (q *JobQueue) setStopped(s bool) {
	q.stoppedMutex.Lock()
	q.stopped = s
	q.stoppedMutex.Unlock()
}

func (q *JobQueue) getStopped() bool {
	q.stoppedMutex.Lock()
	s := q.stopped
	q.stoppedMutex.Unlock()

	return s
}

func (q *JobQueue) dispatch() {
	q.dispatcherStopped.Add(1)
	for {
		select {
		case job := <-q.internalQueue:
			workerChannel := <-q.readyPool
			workerChannel <- job
		case <-q.quit:
			for i := 0; i < len(q.workers); i++ {
				q.workers[i].Stop()
			}
			q.workersStopped.Wait()
			q.dispatcherStopped.Done()
			return
		}
	}
}

func (q *JobQueue) Submit(job Job) {
	if !q.getStopped() {
		q.internalQueue <- job
	}
}
