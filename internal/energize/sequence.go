package energize

import (
	"context"
	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
	"example.com/railvolt/internal/switchgear"
	"fmt"
	"sync"
)

type Operation struct {
	PermitID string          `json:"permit_id"`
	Commands []model.Command `json:"commands"`
	Next     int             `json:"next"`
	Complete bool            `json:"complete"`
	Failed   string          `json:"failed,omitempty"`
}

type Sequencer struct {
	mu         sync.Mutex
	queue      *switchgear.Queue
	client     *switchgear.Client
	permits    *permit.Service
	alarms     *alarm.Service
	operations map[string]Operation
}

func NewSequencer(q *switchgear.Queue, c *switchgear.Client, p *permit.Service, a *alarm.Service) *Sequencer {
	return &Sequencer{queue: q, client: c, permits: p, alarms: a, operations: map[string]Operation{}}
}

func (s *Sequencer) Plan(permitID string, devices []string) Operation {
	s.mu.Lock()
	defer s.mu.Unlock()
	op := Operation{PermitID: permitID}
	for _, device := range devices {
		op.Commands = append(op.Commands, s.queue.Plan(permitID, device, "close"))
	}
	s.operations[permitID] = op
	return op
}

func (s *Sequencer) Step(ctx context.Context, permitID string) (Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.operations[permitID]
	if !ok {
		return Operation{}, fmt.Errorf("energize operation not found")
	}
	if op.Failed != "" || op.Complete {
		return op, nil
	}
	select {
	case <-ctx.Done():
		op.Failed = ctx.Err().Error()
		s.operations[permitID] = op
		return op, ctx.Err()
	default:
	}
	if op.Next >= len(op.Commands) {
		op.Complete = true
		s.operations[permitID] = op
		return op, nil
	}
	cmd := op.Commands[op.Next]
	if _, err := s.queue.Begin(cmd.ID); err != nil {
		return op, err
	}
	if err := s.client.Execute(cmd); err != nil {
		s.queue.Fail(cmd.ID, err.Error())
		op.Failed = err.Error()
		s.permits.Rollback(permitID, err.Error())
		s.operations[permitID] = op
		return op, err
	}
	confirmed, err := s.queue.Confirm(cmd.ID)
	if err != nil {
		return op, err
	}
	op.Commands[op.Next] = confirmed
	op.Next++
	if op.Next == len(op.Commands) {
		op.Complete = true
		if p, found := s.permits.Get(permitID); found {
			p.EnergizeLocked = false
			p.Phase = model.PhaseReadyToEnergize
			s.permits.Save(p)
			s.permits.Advance(permitID, model.PhaseEnergized)
		}
	}
	s.operations[permitID] = op
	return op, nil
}

func (s *Sequencer) Get(permitID string) (Operation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.operations[permitID]
	op.Commands = append([]model.Command(nil), op.Commands...)
	return op, ok
}
