package permit

import (
	"example.com/railvolt/internal/model"
	"fmt"
	"time"
)

func (s *Service) Rollback(id, reason string) (model.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[id]
	if !ok {
		return model.Permit{}, fmt.Errorf("permit not found")
	}
	p.Phase = model.PhaseFailed
	p.FailureReason = reason
	p.GroundLocked = true
	p.EnergizeLocked = true
	p.UpdatedAt = time.Now().UTC()
	s.permits[id] = p
	return p, nil
}

func (s *Service) UnlockForWork(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[id]
	if !ok {
		return fmt.Errorf("permit not found")
	}
	p.GroundLocked = false
	p.Phase = model.PhaseWorking
	s.permits[id] = p
	return nil
}

func (s *Service) LockEnergize(id string, locked bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[id]
	if !ok {
		return fmt.Errorf("permit not found")
	}
	p.EnergizeLocked = locked
	s.permits[id] = p
	return nil
}
