package executor

import "context"

type Callable interface {
	Call(ctx context.Context) *GPResult
}

type Runnable interface {
	Run(ctx context.Context)
}

type Future interface {
	Get() *GPResult
	IsCancelled() bool
	Cancel() bool
	IsDone() bool
}

type ExecutorService interface {
	// no longer accept new tasks
	Shutdown()
	Submit(task Callable) (Future, error)
	IsShutdown() bool
	// Wait for all the tasks to be completed
	WaitTerminate()
	TaskQueueCap() int
	TaskQueueLength() int
	Cancel() bool
	CurrentGCount() int
}
