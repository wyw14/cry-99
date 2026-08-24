package switchgear

import (
	"example.com/railvolt/internal/model"
	"sync"
)

type EvidenceStore struct {
	mu    sync.RWMutex
	items map[string]model.Command
}

func NewEvidenceStore() *EvidenceStore       { return &EvidenceStore{items: map[string]model.Command{}} }
func (s *EvidenceStore) Put(c model.Command) { s.mu.Lock(); defer s.mu.Unlock(); s.items[c.ID] = c }
func (s *EvidenceStore) Get(id string) (model.Command, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.items[id]
	return c, ok
}
func (s *EvidenceStore) All() []model.Command {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Command, 0, len(s.items))
	for _, c := range s.items {
		out = append(out, c)
	}
	return out
}
