package isolation

import (
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
	"fmt"
)

type Recovery struct {
	PlanID    string `json:"plan_id"`
	Confirmed []Step `json:"confirmed"`
	InFlight  []Step `json:"in_flight"`
	Pending   []Step `json:"pending"`
}

func (s *Service) Cancel(planID string, permits *permit.Service) (Recovery, error) {
	p, ok := s.plans.Get(planID)
	if !ok {
		return Recovery{}, fmt.Errorf("isolation plan not found")
	}
	p.Cancelled = true
	recovery := Recovery{PlanID: p.ID}
	confirmedIDs, movingIDs, pendingIDs := []string{}, []string{}, []string{}
	for _, step := range p.Steps {
		switch step.State {
		case model.CommandConfirmed:
			recovery.Confirmed = append(recovery.Confirmed, step)
			confirmedIDs = append(confirmedIDs, step.DeviceID)
		case model.CommandAccepted, model.CommandMoving:
			recovery.InFlight = append(recovery.InFlight, step)
			movingIDs = append(movingIDs, step.DeviceID)
		default:
			recovery.Pending = append(recovery.Pending, step)
			pendingIDs = append(pendingIDs, step.DeviceID)
		}
	}
	if _, err := permits.Cancel(p.PermitID, confirmedIDs, movingIDs, pendingIDs); err != nil {
		return Recovery{}, err
	}
	s.queue.CancelPending()
	s.plans.Save(p)
	return recovery, nil
}
