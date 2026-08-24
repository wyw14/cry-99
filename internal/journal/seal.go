package journal

import (
	"example.com/railvolt/internal/model"
	"sync"
)

type Sealer struct {
	mu     sync.Mutex
	latest *model.Snapshot
}

func NewSealer() *Sealer { return &Sealer{} }

func (s *Sealer) Seal(snapshot model.Snapshot) model.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot.Committed = true
	s.latest = &snapshot
	return snapshot
}

func (s *Sealer) Latest() (model.Snapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.latest == nil {
		return model.Snapshot{}, false
	}
	return *s.latest, true
}
