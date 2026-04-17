package runtime

import "sync/atomic"

type DrainState struct {
	draining       atomic.Bool
	activeRequests atomic.Int64
}

func (s *DrainState) BeginDrain() {
	if s == nil {
		return
	}
	s.draining.Store(true)
}

func (s *DrainState) Draining() bool {
	if s == nil {
		return false
	}
	return s.draining.Load()
}

func (s *DrainState) StartRequest() bool {
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

func (s *DrainState) FinishRequest() {
	if s == nil {
		return
	}
	s.activeRequests.Add(-1)
}

func (s *DrainState) ActiveRequests() int64 {
	if s == nil {
		return 0
	}
	return s.activeRequests.Load()
}
