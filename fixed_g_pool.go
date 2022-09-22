package executor

import (
	"context"
	slog "github.com/vearne/simplelog"
	"sync"
)

type FixedGPoolOption struct {
	taskQueueCap int
}

type option func(*FixedGPoolOption)

// Optional parameters
func WithTaskQueueCap(taskQueueCap int) option {
	return func(t *FixedGPoolOption) {
		t.taskQueueCap = taskQueueCap
	}
}

type FixedGPool struct {
	wg sync.WaitGroup

	Size int
	// task queue
	TaskChan   chan *FutureTask
	isShutdown *AtomicBool
	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewFixedGPool(ctx context.Context, size int, opts ...option) ExecutorService {
	// check params
	if size <= 0 {
		size = 1
	}

	defaultOpts := &FixedGPoolOption{
		taskQueueCap: SIZE,
	}
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(defaultOpts)
	}

	pool := FixedGPool{}
	pool.Size = size
	pool.ctx, pool.cancel = context.WithCancel(ctx)
	pool.isShutdown = NewAtomicBool(false)
	pool.TaskChan = make(chan *FutureTask, defaultOpts.taskQueueCap)
	for i := 0; i < size; i++ {
		go pool.Consume()
	}
	return &pool
}

func (p *FixedGPool) Cancel() bool {
	p.cancel()
	return true
}

func (p *FixedGPool) TaskQueueCap() int {
	return cap(p.TaskChan)
}

func (p *FixedGPool) TaskQueueLength() int {
	return len(p.TaskChan)
}

func (p *FixedGPool) Consume() {
	for task := range p.TaskChan {
		task.run()
		p.wg.Done()
	}
}

// When submitting tasks, blocking may occur
func (p *FixedGPool) Submit(task Callable) (Future, error) {
	if p.IsShutdown() {
		return nil, PoolShutdownErr
	}
	p.wg.Add(1)
	t := NewFutureTask(p.ctx, task)
	p.TaskChan <- t
	slog.Debug("add task to TaskChan")
	return t, nil
}

func (p *FixedGPool) Shutdown() {
	p.isShutdown.Set(true)
}

func (p *FixedGPool) IsShutdown() bool {
	return p.isShutdown.IsTrue()
}

func (p *FixedGPool) WaitTerminate() {
	if !p.IsShutdown() {
		p.Shutdown()
	}
	p.wg.Wait()
}
