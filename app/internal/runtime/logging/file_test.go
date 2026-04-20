package logging

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDailyLogWriterRotatesAtLocalMidnight(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	rotationTick := make(chan time.Time, 1)
	state := struct {
		mu            sync.Mutex
		current       time.Time
		requestedWait time.Duration
	}{
		current: time.Date(2026, 4, 20, 23, 59, 50, 0, location),
	}

	writer := newDailyLogWriterWithClock(
		"api",
		t.TempDir(),
		location,
		func() time.Time {
			state.mu.Lock()
			defer state.mu.Unlock()
			return state.current
		},
		func(wait time.Duration) <-chan time.Time {
			state.mu.Lock()
			state.requestedWait = wait
			state.mu.Unlock()
			return rotationTick
		},
	)
	defer writer.Close()

	if _, err := writer.Write([]byte("before midnight\n")); err != nil {
		t.Fatalf("write before midnight: %v", err)
	}

	beforePath := filepath.Join(writer.dir, "api-2026-04-20.log")
	if _, err := os.Stat(beforePath); err != nil {
		t.Fatalf("expected pre-rotation log file: %v", err)
	}
	state.mu.Lock()
	requestedWait := state.requestedWait
	state.mu.Unlock()
	if requestedWait != 10*time.Second {
		t.Fatalf("rotation wait = %s, want 10s", requestedWait)
	}

	state.mu.Lock()
	state.current = time.Date(2026, 4, 21, 0, 0, 1, 0, location)
	current := state.current
	state.mu.Unlock()
	rotationTick <- current

	deadline := time.Now().Add(2 * time.Second)
	afterPath := filepath.Join(writer.dir, "api-2026-04-21.log")
	for {
		if _, err := os.Stat(afterPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected rotated log file %s to be created", afterPath)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestNextDailyRotationDelayUsesConfiguredLocation(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	now := time.Date(2026, 4, 20, 23, 0, 0, 0, location)

	wait := nextDailyRotationDelay(now, location)
	if wait != time.Hour {
		t.Fatalf("rotation wait = %s, want 1h", wait)
	}
}
