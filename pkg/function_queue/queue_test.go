package function_queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFunctionQueue_Wait(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		queue := NewFunctionQueue[int, int]()
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
			assert.Nil(t, v.err)
			resultValues[i] = v.r
		}

		for i := 0; i < jobCount; i++ {
			assert.Contains(t, resultValues, i*i)
		}
	})

}
