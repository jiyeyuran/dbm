package dbm

import (
	"context"
)

// Instrumenter defines function type that can be used for instrumetation.
// This function should return a function with no argument as a callback for finished execution.
type Instrumenter func(ctx context.Context, op string, message string, args ...any) func(err error)

// Observe operation.
func (i Instrumenter) Observe(ctx context.Context, op string, message string, args ...any) func(err error) {
	if i != nil {
		return i(ctx, op, message, args)
	}

	return func(err error) {}
}
