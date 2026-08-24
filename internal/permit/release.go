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
	if !ok {
		return model.Permit{}, fmt.Errorf("permit not found")
	}
	// Only a successor of the SAME crew chain may supersede the referenced
	// permit: strictly newer generation AND the same owner. A co-resident
	// permit from a different shift (e.g. a new permit created after
	// handover) must never be selected as the release target, otherwise a
	// stale receipt from the prior shift would advance the new shift's
	// permit. The owner/generation check below in ValidateReceipt is the
	// authoritative guard; this loop merely follows same-owner successors.
	for _, candidate := range s.permits {
		if candidate.Worksite == receipt.Worksite && candidate.Owner == p.Owner && candidate.Generation > p.Generation {
			p = candidate
		}
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
