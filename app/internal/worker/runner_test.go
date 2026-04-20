package worker

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestTickerPassesCancelableContextToHandler(t *testing.T) {
	handlerStarted := make(chan struct{}, 1)
	handlerCanceled := make(chan error, 1)
	worker := NewTicker(TickerOptions{
		Interval: 1 * time.Millisecond,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Handler: JobHandlerFunc(func(ctx context.Context) error {
			handlerStarted <- struct{}{}
			<-ctx.Done()
			handlerCanceled <- ctx.Err()
			return nil
		}),
	})

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

func TestTickerDoesNotEmitIdleDebugNoise(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	worker := NewTicker(TickerOptions{
		Interval: 1 * time.Millisecond,
		Logger:   logger,
		Handler: JobHandlerFunc(func(ctx context.Context) error {
			_ = ctx
			return nil
		}),
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Run(ctx)
	}()

	time.Sleep(5 * time.Millisecond)
	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(output.String()) != "" {
		t.Fatalf("unexpected idle worker logs: %s", output.String())
	}
}

type JobHandlerFunc func(context.Context) error

func (f JobHandlerFunc) Handle(ctx context.Context) error {
	return f(ctx)
}
