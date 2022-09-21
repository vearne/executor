package executor

import "context"

type GPResult struct {
	Value any
	Err   error
}

func NewSingleGPool(ctx context.Context, size int) ExecutorService {
	return NewFixedGPool(ctx, 1)
}
