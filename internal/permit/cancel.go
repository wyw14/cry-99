package permit

import (
	"example.com/railvolt/internal/model"
	"fmt"
	"time"
)

type Cancellation struct {
	PermitID  string
	Confirmed []string
	InFlight  []string
	Pending   []string
	CreatedAt time.Time
}

func (s *Service) Cancel(id string, confirmed, inFlight, pending []string) (Cancellation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.permits[id]
	if !ok {
		return Cancellation{}, fmt.Errorf("permit not found")
	}
	p.Phase = model.PhaseCancelled
	p.GroundLocked = true
	p.EnergizeLocked = true
	p.UpdatedAt = time.Now().UTC()
	s.permits[id] = p
	return Cancellation{PermitID: id, Confirmed: append([]string(nil), confirmed...), InFlight: append([]string(nil), inFlight...), Pending: append([]string(nil), pending...), CreatedAt: time.Now().UTC()}, nil
}
