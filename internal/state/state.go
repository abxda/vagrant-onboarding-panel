// Package state persists the wizard's progress to a JSON file next to the
// binary so a student can close the panel and resume where they left off.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// State is the persisted wizard progress.
type State struct {
	// StepStatus maps a step ID to its last known status string.
	StepStatus map[string]string `json:"stepStatus"`
	// Lang is the UI language ("es" / "en").
	Lang string `json:"lang"`
	// VagrantDir is where the generated Vagrantfile / VM live.
	VagrantDir string `json:"vagrantDir"`

	mu   sync.Mutex
	path string
}

// Load reads the state file, returning a zero-valued (but usable) State if it
// doesn't exist yet.
func Load(path string) *State {
	s := &State{
		StepStatus: map[string]string{},
		Lang:       "es",
		path:       path,
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(b, s)
	if s.StepStatus == nil {
		s.StepStatus = map[string]string{}
	}
	if s.Lang == "" {
		s.Lang = "es"
	}
	s.path = path
	return s
}

// SetStatus records a step's status and persists.
func (s *State) SetStatus(stepID, status string) {
	s.mu.Lock()
	s.StepStatus[stepID] = status
	s.mu.Unlock()
	s.save()
}

// Status returns a step's stored status, or "unknown".
func (s *State) Status(stepID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.StepStatus[stepID]; ok {
		return v
	}
	return "unknown"
}

// SetVagrantDir records the working dir and persists.
func (s *State) SetVagrantDir(dir string) {
	s.mu.Lock()
	s.VagrantDir = dir
	s.mu.Unlock()
	s.save()
}

// Reset clears all progress and persists.
func (s *State) Reset() {
	s.mu.Lock()
	s.StepStatus = map[string]string{}
	s.mu.Unlock()
	s.save()
}

func (s *State) save() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, b, 0o644)
}
