package permit

import (
	"example.com/railvolt/internal/model"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Service struct {
	mu      sync.RWMutex
	permits map[string]model.Permit
}

func NewService() *Service { return &Service{permits: make(map[string]model.Permit)} }

func (s *Service) Create(worksite, owner, topologyRevision string) model.Permit {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := model.Permit{ID: uuid.NewString(), Worksite: worksite, Owner: owner, Generation: 1, Phase: model.PhaseDraft, TopologyRevision: topologyRevision, GroundLocked: true, EnergizeLocked: true, UpdatedAt: time.Now().UTC()}
	s.permits[p.ID] = p
	return p
}

func (s *Service) Save(p model.Permit) error {
	if err := p.Valid(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p.UpdatedAt = time.Now().UTC()
	s.permits[p.ID] = p
	return nil
}

func (s *Service) Get(id string) (model.Permit, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.permits[id]
	return p, ok
}

func (s *Service) List() []model.Permit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Permit, 0, len(s.permits))
	for _, p := range s.permits {
		out = append(out, p)
	}
	return out
}

func (s *Service) Advance(id string, next model.Phase) (model.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[id]
	if !ok {
		return model.Permit{}, fmt.Errorf("permit not found")
	}
	if !validTransition(p.Phase, next) {
		return model.Permit{}, fmt.Errorf("invalid transition %s -> %s", p.Phase, next)
	}
	p.Phase, p.UpdatedAt = next, time.Now().UTC()
	s.permits[id] = p
	return p, nil
}

func validTransition(from, to model.Phase) bool {
	if to == model.PhaseFailed || to == model.PhaseCancelled {
		return from != model.PhaseEnergized
	}
	allowed := map[model.Phase]model.Phase{model.PhaseDraft: model.PhaseIsolating, model.PhaseIsolating: model.PhaseIsolated, model.PhaseIsolated: model.PhaseGrounded, model.PhaseGrounded: model.PhaseWorking, model.PhaseWorking: model.PhaseReleasing, model.PhaseReleasing: model.PhaseReadyToEnergize, model.PhaseReadyToEnergize: model.PhaseEnergized}
	return allowed[from] == to
}
