package xlog

import (
	"context"
	"testing"
)

func TestContextLogger(t *testing.T) {
	ctx := context.Background()
	doSomething(ctx)
}

func doSomething(ctx context.Context) {
	logger := FromContext(ctx)
	logger.Info("hello world")
	logger = logger.WithName("foo")
	ctx = NewContext(ctx, logger)
	doSomeMore(ctx)
}

func doSomeMore(ctx context.Context) {
	logger := FromContext(ctx)
	logger.Info("hello also from me")
}
