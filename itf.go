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
	// 不再接收新的任务，等所有任务执行完毕，就终止协程池
	Shutdown()
	Submit(task Callable) Future
	IsShutdown() bool
	WaitTerminate()
}
