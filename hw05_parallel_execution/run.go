// Package hw05parallelexecution is for homework on OTUS Professional #10
package hw05parallelexecution

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	// Place your code here.
	var wg sync.WaitGroup
	var err error
	taskChan := make(chan Task)
	coordinatorCtx, coordinatorCancelFunc := context.WithCancel(context.Background())
	errChan := make(chan struct{})
	FillChanWithTasks(coordinatorCtx, tasks, &wg, taskChan)
	TaskErrorShutdownCoordinator(coordinatorCancelFunc, &wg, errChan, m)

	for i := range n {
		wg.Go(func() {
			for {
				select {
				case task, ok := <-taskChan:
					if !ok {
						log.Printf("task channel closed, worker #%d shutdown\n", i)
						return
					}
					ProcessTask(i, task, errChan)
				case <-coordinatorCtx.Done():
					err = ErrErrorsLimitExceeded
					log.Printf("task channel closed, worker #%d shutdown\n", i)
					return
				}
			}
		})
	}
	wg.Wait()
	return err
}

// TaskErrorShutdownCoordinator think about correct naming, because now it's bullshit.
func TaskErrorShutdownCoordinator(cancel context.CancelFunc, wg *sync.WaitGroup, errChan chan struct{}, m int) {
	wg.Go(func() {
		count := 0
		for {
			<-errChan
			count++
			if count >= m {
				cancel()
				return
			}
		}
	})
}

// FillChanWithTasks takes our slice of tasks and push it to the channel with non-blocking algo.
func FillChanWithTasks(coordinatorCtx context.Context, tasks []Task, wg *sync.WaitGroup, taskChan chan Task) {
	wg.Go(func() {
		for _, v := range tasks {
			select {
			// While we're trying to push some task in channel -> channel can be already blocked AND we can have many
			// messages with errors
			// So we just create priority select on writing.
			case <-coordinatorCtx.Done():
				close(taskChan)
				return
			default:
				taskChan <- v
			}
		}
		close(taskChan)
	})
}

func ProcessTask(workerNumber int, t Task, errChan chan struct{}) {
	err := t()
	if err != nil {
		fmt.Printf("Occure error in worker #%d: %v\n", workerNumber, err)
		// think about blocking while writing. Is it safe? Maybe I need create some timeout or...?
		errChan <- struct{}{}
	} else {
		fmt.Printf("Worker #%d completed task. Taking new\n", workerNumber)
	}
}
