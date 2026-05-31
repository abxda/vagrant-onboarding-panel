// Package logsink is a small in-memory ring buffer of log lines with a
// subscriber hook, used to stream a step's live output to the Wails frontend.
package logsink

import (
	"sync"
	"time"
)

// Line is one log entry. Time is stamped on the Go side so the frontend
// never has to compute it.
type Line struct {
	Time  string `json:"time"`  // HH:MM:SS
	Level string `json:"level"` // INFO | WARN | ERROR
	Text  string `json:"text"`
}

// Sink is a fixed-capacity ring buffer with an optional emit callback.
type Sink struct {
	mu    sync.Mutex
	buf   []Line
	cap   int
	onNew func(Line)
}

// New returns a Sink holding up to capacity lines.
func New(capacity int) *Sink {
	if capacity <= 0 {
		capacity = 2000
	}
	return &Sink{cap: capacity, buf: make([]Line, 0, capacity)}
}

// OnNew registers a callback fired for every appended line (used to forward
// to Wails events). Only one subscriber; last writer wins.
func (s *Sink) OnNew(fn func(Line)) {
	s.mu.Lock()
	s.onNew = fn
	s.mu.Unlock()
}

// Emit appends a line, trimming the oldest if at capacity, and notifies the
// subscriber outside the lock.
func (s *Sink) Emit(level, text string) {
	ln := Line{Time: time.Now().Format("15:04:05"), Level: level, Text: text}
	s.mu.Lock()
	if len(s.buf) >= s.cap {
		copy(s.buf, s.buf[1:])
		s.buf = s.buf[:s.cap-1]
	}
	s.buf = append(s.buf, ln)
	cb := s.onNew
	s.mu.Unlock()
	if cb != nil {
		cb(ln)
	}
}

// Snapshot returns a copy of the current buffer.
func (s *Sink) Snapshot() []Line {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Line, len(s.buf))
	copy(out, s.buf)
	return out
}

// Clear empties the buffer.
func (s *Sink) Clear() {
	s.mu.Lock()
	s.buf = s.buf[:0]
	s.mu.Unlock()
}
