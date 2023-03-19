package function_queue

import "sync"

const defaultConcurrency = 3

type Function[V, R any] func(V) (R, error)
type queuedFunction[V, R any] struct {
	f Function[V, R]
	v V
}

type Result[R any] struct {
	R   R
	Err error
}

type Runner[V, R any] interface {
	Run(f Function[V, R], v V)
}

type Waiter[V, R any] interface {
	Wait() []Result[R]
}

type FunctionQueue[V, R any] struct {
	queuedFunctions   chan queuedFunction[V, R]
	functionWaitGroup sync.WaitGroup

	results     []Result[R]
	resultMutex sync.Mutex
}

func (queue *FunctionQueue[V, R]) Run(f Function[V, R], v V) {
	queue.functionWaitGroup.Add(1)
	queue.queuedFunctions <- queuedFunction[V, R]{f, v}
}

func (queue *FunctionQueue[V, R]) Wait() []Result[R] {
	queue.functionWaitGroup.Wait()

	close(queue.queuedFunctions)

	return queue.results
}

func NewFunctionQueue[V, R any](concurrency int) *FunctionQueue[V, R] {
	queue := FunctionQueue[V, R]{
		queuedFunctions: make(chan queuedFunction[V, R]),
	}

	if concurrency == 0 {
		concurrency = defaultConcurrency
	}

	for worker := 0; worker < concurrency; worker++ {
		go func() {
			for function := range queue.queuedFunctions {
				result, err := function.f(function.v)

				queue.resultMutex.Lock()
				queue.results = append(queue.results, Result[R]{
					R:   result,
					Err: err,
				})
				queue.resultMutex.Unlock()

				queue.functionWaitGroup.Done()
			}
		}()
	}

	return &queue
}
