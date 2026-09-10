package hw05parallelexecution

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("stops after reaching the error limit", testRunErrorLimit)
	t.Run("returns error for non-positive worker count", testRunNonPositiveWorkers)
	t.Run("returns error for negative error limit", testRunNegativeErrorLimit)
	t.Run("runs tasks concurrently without errors", testRunConcurrentlyWithoutErrors)
}

func testRunErrorLimit(t *testing.T) {
	tasksCount := 50
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32

	for i := 0; i < tasksCount; i++ {
		err := fmt.Errorf("error from task %d", i)
		tasks = append(tasks, func() error {
			rand, _ := rand.Int(rand.Reader, big.NewInt(100))
			time.Sleep(time.Millisecond * time.Duration(rand.Int64()))
			atomic.AddInt32(&runTasksCount, 1)
			return err
		})
	}

	workersCount := 10
	maxErrorsCount := 23
	err := Run(tasks, workersCount, maxErrorsCount)

	require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
	require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
}

func testRunNonPositiveWorkers(t *testing.T) {
	tasks := []Task{func() error { return nil }}

	err := Run(tasks, -5, 10)

	require.ErrorIs(t, err, ErrNegativeOrZeroWorkersCount)
}

func testRunNegativeErrorLimit(t *testing.T) {
	tasks := []Task{func() error { return nil }}

	err := Run(tasks, 5, -10)

	require.ErrorIs(t, err, ErrNegativeMaxErrorCount)
}

func testRunConcurrentlyWithoutErrors(t *testing.T) {
	tasksCount := 50
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32
	var sumTime time.Duration

	for i := 0; i < tasksCount; i++ {
		rand, _ := rand.Int(rand.Reader, big.NewInt(100))
		taskSleep := time.Millisecond * time.Duration(rand.Int64())
		sumTime += taskSleep

		tasks = append(tasks, func() error {
			time.Sleep(taskSleep)
			atomic.AddInt32(&runTasksCount, 1)
			return nil
		})
	}

	workersCount := 5
	maxErrorsCount := 1

	start := time.Now()
	err := Run(tasks, workersCount, maxErrorsCount)
	elapsedTime := time.Since(start)
	require.NoError(t, err)

	require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
}
