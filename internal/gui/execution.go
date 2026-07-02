package gui

import (
	"sync"
)

// ExecutionState holds the thread-safe execution state for async operations.
type ExecutionState struct {
	mu      sync.Mutex
	loading bool
	result  string
	err     string
	running bool
}

// IsLoading returns whether a generation is in progress (thread-safe).
func (s *ExecutionState) IsLoading() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loading
}

// Result returns the current result text (thread-safe).
func (s *ExecutionState) Result() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}

// Error returns the current error text (thread-safe).
func (s *ExecutionState) Error() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// SetLoading sets the loading state (thread-safe).
func (s *ExecutionState) SetLoading(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loading = v
}

// SetResult sets the result text (thread-safe).
func (s *ExecutionState) SetResult(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = v
}

// SetError sets the error text (thread-safe).
func (s *ExecutionState) SetError(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = v
}

// TryAcquire attempts to mark execution as running.
// Returns false if already running.
func (s *ExecutionState) TryAcquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	return true
}

// Release marks execution as not running (thread-safe).
func (s *ExecutionState) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}
