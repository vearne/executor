package executor

import "errors"

var (
	ErrTaskCanceled = errors.New("task has been canceled")
	ErrPoolShutdown = errors.New("pool has been shutdown")
)
