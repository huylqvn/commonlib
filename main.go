package main

import (
	"fmt"
	"time"

	"github.com/huylqvn/commonlib/async"
)

type Job struct {
	Val int
}

func (j Job) Process() {
	fmt.Println(j.Val)
	time.Sleep(2 * time.Second)
}

func NewJob(val int) *Job {
	return &Job{Val: val}
}

func main() {
	queue := async.NewJobQueue(5)

	queue.Start()

	queue.Submit(NewJob(1))
	queue.Submit(NewJob(2))
	queue.Submit(NewJob(3))
	queue.Submit(NewJob(4))
	queue.Submit(NewJob(5))
	queue.Submit(NewJob(6))
	queue.Submit(NewJob(7))
	queue.Submit(NewJob(8))
	queue.Submit(NewJob(9))
	queue.Submit(NewJob(10))

	queue.Stop()
}
