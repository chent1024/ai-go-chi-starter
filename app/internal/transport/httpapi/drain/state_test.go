package drain

import "testing"

func TestStateLifecycle(t *testing.T) {
	var state State

	if !state.StartRequest() {
		t.Fatal("StartRequest() = false, want true before drain")
	}
	if got := state.ActiveRequests(); got != 1 {
		t.Fatalf("ActiveRequests() = %d", got)
	}

	state.FinishRequest()
	state.BeginDrain()

	if state.StartRequest() {
		t.Fatal("StartRequest() = true, want false after drain")
	}
	if !state.Draining() {
		t.Fatal("Draining() = false, want true")
	}
}
