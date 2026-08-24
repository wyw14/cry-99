package crew

import (
	"example.com/railvolt/internal/model"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Session struct {
	ID, Worksite, Owner, Fencing string
	Generation                   uint64
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]Session
	current  map[string]Session
}

func NewManager() *Manager {
	return &Manager{sessions: map[string]Session{}, current: map[string]Session{}}
}

func (m *Manager) Handover(worksite, owner string, generation uint64) Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Session{ID: uuid.NewString(), Worksite: worksite, Owner: owner, Fencing: uuid.NewString(), Generation: generation}
	m.sessions[s.ID] = s
	m.current[worksite] = s
	return s
}

func (m *Manager) Current(worksite string) (Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.current[worksite]
	return s, ok
}

func (m *Manager) Receipt(s Session, permitID string) model.Receipt {
	return model.Receipt{PermitID: permitID, Worksite: s.Worksite, Owner: s.Owner, Generation: s.Generation, Fencing: s.Fencing}
}

func (m *Manager) ValidateReceipt(r model.Receipt, permit model.Permit) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	current, ok := m.current[permit.Worksite]
	if !ok || current.Worksite != r.Worksite {
		return fmt.Errorf("crew receipt fenced")
	}
	// A release receipt may only act on a permit owned by the crew currently
	// in charge of the worksite. A stale receipt from a prior shift (different
	// owner/generation) must not mutate the permit that the new shift owns,
	// even if it references the same worksite or permit id.
	if current.Owner != permit.Owner {
		return fmt.Errorf("permit owner mismatch expected=%s actual=%s", current.Owner, permit.Owner)
	}
	if r.Owner != permit.Owner || r.Generation != permit.Generation {
		return fmt.Errorf("permit owner mismatch expected=%s actual=%s", permit.Owner, r.Owner)
	}
	return nil
}
