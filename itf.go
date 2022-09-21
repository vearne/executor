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
	Submit(task Callable) Future
	IsShutdown() bool
	// 等所有任务执行完毕，就终止协程池
	WaitTerminate()
	TaskQueueCap() int
	TaskQueueLength() int
}
