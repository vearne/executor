package executor

import (
	"context"
)

type FutureTask struct {
	c           Callable
	ch          chan *GPResult
	isDone      *AtomicBool
	isCancelled *AtomicBool
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewFutureTask(ctx context.Context, c Callable) *FutureTask {
	t := FutureTask{}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.c = c
	t.ch = make(chan *GPResult, 1)
	t.isDone = NewAtomicBool(false)
	t.isCancelled = NewAtomicBool(false)
	return &t
}

func (f *FutureTask) Get() *GPResult {
	result := <-f.ch
	return result
}

func (f *FutureTask) IsCancelled() bool {
	return f.isCancelled.IsTrue()
}

func (f *FutureTask) Cancel() bool {
	f.isCancelled.Set(true)
	f.cancel()
	return true
}

func (f *FutureTask) run() {
	if f.IsCancelled() {
		return
	}
	select {
	case f.ch <- f.c.Call(f.ctx):
		f.isDone.Set(true)
	case <-f.ctx.Done():
		// cancel
		f.ch <- &GPResult{Err: ErrTaskCanceled}
	}
}

func (f *FutureTask) IsDone() bool {
	return f.isDone.IsTrue()
}
