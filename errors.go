package executor

import "errors"

var (
	TaskCanceledErr = errors.New("task has been canceled")
	PoolShutdownErr = errors.New("pool has been shutdown")
)
