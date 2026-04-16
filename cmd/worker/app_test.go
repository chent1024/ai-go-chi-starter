package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestTickerWorkerPassesCancelableContextToHandler(t *testing.T) {
	handlerStarted := make(chan struct{}, 1)
	handlerCanceled := make(chan error, 1)
	worker := tickerWorker{
		interval: 1 * time.Millisecond,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		handler: JobHandlerFunc(func(ctx context.Context) error {
			handlerStarted <- struct{}{}
			<-ctx.Done()
			handlerCanceled <- ctx.Err()
			return nil
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Run(ctx)
	}()

	select {
	case <-handlerStarted:
	case <-time.After(1 * time.Second):
		t.Fatal("handler did not start")
	}
	cancel()

	select {
	case err := <-handlerCanceled:
		if err != context.Canceled {
			t.Fatalf("handler ctx err = %v, want context.Canceled", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("handler did not observe cancellation")
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

type JobHandlerFunc func(context.Context) error

func (f JobHandlerFunc) Handle(ctx context.Context) error {
	return f(ctx)
}
