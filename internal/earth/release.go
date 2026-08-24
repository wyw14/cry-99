package earth

import (
	"example.com/railvolt/internal/model"
	"fmt"
)

func (s *Service) Release(id string) (model.TelemetryEvidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.devices[id]
	if !ok {
		return model.TelemetryEvidence{}, fmt.Errorf("ground not found")
	}
	samples := s.collector.Samples(e.AreaID)
	durable, err := s.evidence.Save(e.AreaID, samples, s.collector.Zero(e.AreaID))
	if err != nil {
		return e, err
	}
	if !durable.SafeToRelease() {
		return e, fmt.Errorf("telemetry evidence is not safe to release")
	}
	e.ZeroCurrent, e.Durable = durable.ZeroCurrent, durable.Durable
	s.devices[id] = e
	return e, nil
}

func (s *Service) Released(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.devices[id]
	return ok && e.Durable && e.ZeroCurrent
}
