package async

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestFunctionQueue_Run(t *testing.T) {
	t.Run("Jobs start execution before Wait() is called", func(t *testing.T) {
		queue := NewFunctionQueue[int, int](3)
		lock := sync.Mutex{}

		queue.Run(func(_ int) (int, error) {
			if !lock.TryLock() {
				t.Fatal("Queue did not start processing until after Wait() was called")
			}

			lock.Unlock()

			return 1, nil
		}, 0)

		time.Sleep(100 * time.Millisecond)
		lock.Lock()

		_ = queue.Wait()
	})
}

func TestFunctionQueue_Wait(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		queue := NewFunctionQueue[int, int](3)
		jobCount := 10

		for i := 0; i < jobCount; i++ {
			queue.Run(func(i int) (int, error) {
				return i * i, nil
			}, i)
		}

		results := queue.Wait()

		assert.Len(t, results, jobCount)

		resultValues := make([]int, len(results))

		for i, v := range results {
			assert.Nil(t, v.Err)
			resultValues[i] = v.R
		}

		for i := 0; i < jobCount; i++ {
			assert.Contains(t, resultValues, i*i)
		}
	})
}

func TestNewFunctionQueue(t *testing.T) {
	t.Run("Default concurrency is 1", func(t *testing.T) {
		queue := NewFunctionQueue[int, int](0)
		lock := sync.Mutex{}

		queue.Run(func(_ int) (int, error) {
			lock.Lock()
			time.Sleep(100 * time.Millisecond)
			lock.Unlock()

			return 1, nil
		}, 0)

		queue.Run(func(_ int) (int, error) {
			if !lock.TryLock() {
				t.Fatal("Tried to run two functions at the same time")
			}

			lock.Unlock()

			return 3, nil
		}, 2)

		_ = queue.Wait()
	})
}

func BenchmarkFunctionQueue_Wait(b *testing.B) {
	for i := 0; i <= 10; i++ {
		concurrency := i * 10

		b.Run(fmt.Sprintf("Concurrency of %d", concurrency), func(b *testing.B) {
			b.Run("async package", func(b *testing.B) {
				asyncMultiplication(concurrency, b.N)
			})

			b.Run("workerpool package", func(b *testing.B) {
				workerpoolMultiplication(concurrency, b.N)
			})
		})
	}
}

func asyncMultiplication(concurrency, count int) {
	multiplierQueue := NewFunctionQueue[int, int](concurrency)

	for i := 0; i < count; i++ {
		multiplierQueue.Run(func(i int) (int, error) {
			return i * i, nil
		}, i)
	}

	results := multiplierQueue.Wait()

	for _, v := range results {
		if v.Err != nil {
			panic(fmt.Sprintf("Error encountered while benchmarking: %v", v.Err))
		}
	}
}

func workerpoolMultiplication(concurrency, count int) {
	multiplierQueue := workerpool.New(concurrency)
	results := make([]int, count)

	for i := 0; i < count; i++ {
		i := i

		multiplierQueue.Submit(func() {
			results[i] = i * i
		})
	}

	multiplierQueue.StopWait()
}
