package executor

import (
	"context"
	"sync"
)

const (
	SIZE int = 50
)

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

func (p *FixedGPool) TaskQueueCap() int {
	return cap(p.TaskChan)
}

func (p *FixedGPool) TaskQueueLength() int {
	return len(p.TaskChan)
}

func NewFixedGPool(ctx context.Context, size int) ExecutorService {
	pool := FixedGPool{}
	pool.Size = size
	pool.ctx, pool.cancel = context.WithCancel(ctx)
	pool.isShutdown = NewAtomicBool(false)
	pool.TaskChan = make(chan *FutureTask, SIZE)
	for i := 0; i < size; i++ {
		go pool.Consume()
	}
	return &pool
}

func (p *FixedGPool) Consume() {
	for task := range p.TaskChan {
		task.run()
		p.wg.Done()
	}
}

func (p *FixedGPool) Submit(task Callable) Future {
	t := NewFutureTask(p.ctx, task)
	p.TaskChan <- t
	p.wg.Add(1)
	return t
}

func (p *FixedGPool) Shutdown() {
	close(p.TaskChan)
	p.isShutdown.Set(true)
}

func (p *FixedGPool) IsShutdown() bool {
	return p.isShutdown.IsTrue()
}

func (p *FixedGPool) WaitTerminate() {
	if !p.IsShutdown() {
		panic("pool must shutdown first!")
	}
	p.wg.Wait()
}
