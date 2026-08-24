package isolation

import (
	"context"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
	"fmt"
)

type Service struct {
	plans *Registry
	queue *switchgear.Queue
}

func NewService(plans *Registry, queue *switchgear.Queue) *Service {
	return &Service{plans: plans, queue: queue}
}

func (s *Service) Begin(permitID string, devices []string) Plan {
	p := s.plans.Create(permitID, devices)
	for i := range p.Steps {
		c := s.queue.Plan(permitID, p.Steps[i].DeviceID, p.Steps[i].Action)
		p.Steps[i].CommandID = c.ID
	}
	s.plans.Save(p)
	return p
}

func (s *Service) ExecuteNext(ctx context.Context, planID string) (Plan, error) {
	p, ok := s.plans.Get(planID)
	if !ok {
		return Plan{}, fmt.Errorf("isolation plan not found")
	}
	if p.Cancelled {
		return p, fmt.Errorf("isolation cancelled")
	}
	if p.Current >= len(p.Steps) {
		return p, nil
	}
	select {
	case <-ctx.Done():
		return p, ctx.Err()
	default:
	}
	step := &p.Steps[p.Current]
	if _, err := s.queue.Begin(step.CommandID); err != nil {
		return p, err
	}
	step.State = model.CommandAccepted
	if _, err := s.queue.Move(step.CommandID); err != nil {
		return p, err
	}
	step.State = model.CommandMoving
	s.plans.Save(p)
	return p, nil
}

func (s *Service) Confirm(planID string) (Plan, error) {
	p, ok := s.plans.Get(planID)
	if !ok {
		return Plan{}, fmt.Errorf("isolation plan not found")
	}
	if p.Current >= len(p.Steps) {
		return p, nil
	}
	step := &p.Steps[p.Current]
	if _, err := s.queue.Confirm(step.CommandID); err != nil {
		return p, err
	}
	step.State = model.CommandConfirmed
	p.Current++
	s.plans.Save(p)
	return p, nil
}

func (s *Service) Plan(planID string) (Plan, bool) { return s.plans.Get(planID) }
