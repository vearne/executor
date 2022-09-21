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
	Ctx context.Context
}

func NewFixedGPool(ctx context.Context, size int) ExecutorService {
	pool := FixedGPool{}
	pool.Size = size
	pool.Ctx = ctx
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
	t := NewFutureTask(p.Ctx, task)
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
