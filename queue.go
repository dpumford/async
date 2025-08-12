package async

import (
	"context"
	"sync"
)

const defaultConcurrency = 1

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

	ctx           context.Context
	cancel        context.CancelFunc
	canceled      bool
	canceledMutex sync.Mutex
}

func (queue *FunctionQueue[V, R]) Run(f Function[V, R], v V) {
	queue.canceledMutex.Lock()
	defer queue.canceledMutex.Unlock()
	if queue.canceled {
		return
	}

	queue.functionWaitGroup.Add(1)
	queue.queuedFunctions <- queuedFunction[V, R]{f, v}
}

func (queue *FunctionQueue[V, R]) Wait() []Result[R] {
	queue.functionWaitGroup.Wait()

	queue.canceledMutex.Lock()
	if !queue.canceled {
		close(queue.queuedFunctions)
	}
	queue.canceledMutex.Unlock()

	return queue.results
}

func NewFunctionQueue[V, R any](ctx context.Context, concurrency int) *FunctionQueue[V, R] {
	innerCtx, cancel := context.WithCancel(ctx)

	queue := FunctionQueue[V, R]{
		queuedFunctions: make(chan queuedFunction[V, R]),
		ctx:             innerCtx,
		cancel:          cancel,
	}

	if concurrency == 0 {
		concurrency = defaultConcurrency
	}

	go func() {
		<-queue.ctx.Done()
		queue.canceledMutex.Lock()
		queue.canceled = true
		close(queue.queuedFunctions)
		queue.canceledMutex.Unlock()
	}()

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
