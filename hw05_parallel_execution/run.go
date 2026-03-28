package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, workersCount, maxErrorsCount int) error {
	if workersCount <= 0 || maxErrorsCount <= 0 {
		return ErrErrorsLimitExceeded
	}

	tasksCh := make(chan Task)
	var wg sync.WaitGroup

	var errCount int

	var mtx sync.Mutex
	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				if err := task(); err != nil {
					mtx.Lock()
					errCount++
					mtx.Unlock()
				}
			}
		}()
	}

	for _, task := range tasks {
		mtx.Lock()
		if errCount >= maxErrorsCount {
			mtx.Unlock()
			break
		}
		mtx.Unlock()
		tasksCh <- task
	}

	close(tasksCh)
	wg.Wait()

	if errCount >= maxErrorsCount {
		return ErrErrorsLimitExceeded
	}

	return nil
}
