package executor

import "context"

type GPResult struct {
	Value any
	Err   error
}

func NewSingleGPool(ctx context.Context, opts ...option) ExecutorService {
	return NewFixedGPool(ctx, 1, opts...)
}
