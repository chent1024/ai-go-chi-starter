package drain

import "sync/atomic"

type State struct {
	draining       atomic.Bool
	activeRequests atomic.Int64
}

func (s *State) BeginDrain() {
	if s == nil {
		return
	}
	s.draining.Store(true)
}

func (s *State) Draining() bool {
	if s == nil {
		return false
	}
	return s.draining.Load()
}

func (s *State) StartRequest() bool {
	if s == nil {
		return true
	}
	if s.draining.Load() {
		return false
	}
	s.activeRequests.Add(1)
	if s.draining.Load() {
		s.activeRequests.Add(-1)
		return false
	}
	return true
}

func (s *State) FinishRequest() {
	if s == nil {
		return
	}
	s.activeRequests.Add(-1)
}

func (s *State) ActiveRequests() int64 {
	if s == nil {
		return 0
	}
	return s.activeRequests.Load()
}
