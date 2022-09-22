package executor

import "errors"

var (
	TaskCanceledErr = errors.New("task has been canceled")
)
