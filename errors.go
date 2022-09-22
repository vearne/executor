package executor

import "errors"

var (
	ErrTaskCanceled        = errors.New("task has been canceled")
	ErrInvalidTaskQueueCap = errors.New("invalid taskQueueCap for pool")
	ErrShutDowned          = errors.New("pool has been shutdown")
)
