// Package hw05parallelexecution is for homework on OTUS Professional #10
package hw05parallelexecution

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	// ErrErrorsLimitExceeded too many errors.
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	// ErrNegativeOrZeroWorkersCount negative workers number (or zero).
	ErrNegativeOrZeroWorkersCount = errors.New("workers number less or equal zero")
	// ErrNegativeMaxErrorCount allowable count of errors is negative.
	ErrNegativeMaxErrorCount = errors.New("count of errors can't be less than zero")
)

// Task represent some work for to do.
type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	if n <= 0 {
		log.Println("numbers of workers can't be less or equal zero")
		return ErrNegativeOrZeroWorkersCount
	}

	if m < 0 {
		log.Println("count of errors can't be less then zero")
		return ErrNegativeMaxErrorCount
	}

	allTasksCount := &AllTasksCounter{all: len(tasks)}
	var wg sync.WaitGroup
	taskChan := make(chan Task)
	coordinatorCtx, coordinatorCancelFunc := context.WithCancel(context.Background())
	successChan := make(chan bool)
	FillChanWithTasks(coordinatorCtx, tasks, &wg, taskChan)
	hasCompletedNormally := TaskErrorShutdownCoordinator(coordinatorCancelFunc, &wg, successChan, m, allTasksCount)
	for i := range n {
		wg.Go(func() {
			for {
				select {
				case task, ok := <-taskChan:
					if !ok {
						log.Printf("task channel closed, worker #%d shutdown\n", i)
						return
					}
					ProcessTask(coordinatorCtx, i, task, successChan)
				case <-coordinatorCtx.Done():
					return
				}
			}
		})
	}
	err := <-hasCompletedNormally
	wg.Wait()
	if err != nil {
		return fmt.Errorf("reached maximum errors: %w", err)
	}
	return err
}

// TaskErrorShutdownCoordinator think about correct naming, because now it's bullshit.
func TaskErrorShutdownCoordinator(
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
	successChan chan bool,
	m int,
	allTasks *AllTasksCounter,
) chan error {
	hasCompletedNormally := make(chan error)
	wg.Go(func() {
		for success := range successChan {
			if !success {
				// if we expect no error, but got one
				if m == 0 {
					log.Println("reached maximum error's count. Canceling work")
					hasCompletedNormally <- ErrErrorsLimitExceeded
					cancel()
					return
				}
				allTasks.IncrementErrorCounter()
			} else {
				allTasks.IncrementSuccessCounter()
			}
			if allTasks.GetAllErrorsCount() == m && m != 0 {
				log.Println("reached maximum error's count. Canceling work")
				hasCompletedNormally <- ErrErrorsLimitExceeded
				cancel()
				return
			}
			if (allTasks.GetAllErrorsCount() + allTasks.GetAllSuccessCount()) == allTasks.GetAll() {
				hasCompletedNormally <- nil
				cancel()
				return
			}
		}
	})
	return hasCompletedNormally
}

// FillChanWithTasks takes our slice of tasks and push it to the channel with non-blocking algo.
func FillChanWithTasks(coordinatorCtx context.Context, tasks []Task, wg *sync.WaitGroup, taskChan chan Task) {
	wg.Go(func() {
		for _, v := range tasks {
			select {
			case <-coordinatorCtx.Done():
				close(taskChan)
				return
			case taskChan <- v:
			}
		}
		close(taskChan)
	})
}

// ProcessTask is calling task with some wrapping.
func ProcessTask(ctx context.Context, workerNumber int, t Task, successChan chan bool) {
	success := true
	err := t()
	if err != nil {
		fmt.Printf("Occure error in worker #%d: %v\n", workerNumber, err)
		success = false
	}
	fmt.Printf("Worker #%d completed task. Taking new\n", workerNumber)
	select {
	case successChan <- success:
	case <-ctx.Done():
	}
}

// AllTasksCounter represent struct for counting state of remain/all tasks.
type AllTasksCounter struct {
	all            int
	errorsCounter  int
	successCounter int
}

// GetAllSuccessCount provide count of all success completed task.
func (t *AllTasksCounter) GetAllSuccessCount() int {
	return t.successCounter
}

// GetAllErrorsCount provide of all success completed task.
func (t *AllTasksCounter) GetAllErrorsCount() int {
	return t.errorsCounter
}

// IncrementSuccessCounter increment count of completed tasks.
func (t *AllTasksCounter) IncrementSuccessCounter() {
	t.successCounter++
}

// IncrementErrorCounter increment count of tasks with errors.
func (t *AllTasksCounter) IncrementErrorCounter() {
	t.errorsCounter++
}

// GetAll provide us count of all given tasks.
func (t *AllTasksCounter) GetAll() int {
	return t.all
}
