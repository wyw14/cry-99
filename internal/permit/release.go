package permit

import (
	"example.com/railvolt/internal/crew"
	"example.com/railvolt/internal/model"
	"fmt"
	"time"
)

func (s *Service) AcceptRelease(m *crew.Manager, receipt model.Receipt) (model.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[receipt.PermitID]
	for _, candidate := range s.permits {
		if candidate.Worksite == receipt.Worksite && candidate.Generation >= p.Generation {
			p, ok = candidate, true
		}
	}
	if !ok {
		return model.Permit{}, fmt.Errorf("permit not found")
	}
	if err := m.ValidateReceipt(receipt, p); err != nil {
		return model.Permit{}, err
	}
	if p.Phase != model.PhaseReleasing && p.Phase != model.PhaseWorking {
		return model.Permit{}, fmt.Errorf("permit is not releasable")
	}
	p.Phase = model.PhaseReadyToEnergize
	p.UpdatedAt = now()
	s.permits[p.ID] = p
	return p, nil
}

func now() time.Time { return time.Now().UTC() }
