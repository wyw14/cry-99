package journal

import (
	"encoding/json"
	"example.com/railvolt/internal/model"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu       sync.Mutex
	path     string
	events   []map[string]any
	revision uint64
}

func NewStore(path string) *Store { return &Store{path: path, events: []map[string]any{}} }

func (s *Store) Append(kind string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revision++
	s.events = append(s.events, map[string]any{"revision": s.revision, "kind": kind, "value": value})
	return s.persistLocked()
}

func (s *Store) Events() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]any, len(s.events))
	copy(out, s.events)
	return out
}

func (s *Store) Revision() uint64 { s.mu.Lock(); defer s.mu.Unlock(); return s.revision }

func (s *Store) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		return nil
	}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &s.events); err != nil {
		return fmt.Errorf("journal decode: %w", err)
	}
	for _, event := range s.events {
		if n, ok := event["revision"].(float64); ok && uint64(n) > s.revision {
			s.revision = uint64(n)
		}
	}
	return nil
}

func (s *Store) SaveSnapshot(snapshot model.Snapshot) error { return s.Append("snapshot", snapshot) }
