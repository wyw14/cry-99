package earth

import (
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/telemetry"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Service struct {
	mu        sync.RWMutex
	devices   map[string]model.TelemetryEvidence
	collector *telemetry.Collector
	evidence  *telemetry.EvidenceStore
}

func NewService(collector *telemetry.Collector, evidence *telemetry.EvidenceStore) *Service {
	return &Service{devices: map[string]model.TelemetryEvidence{}, collector: collector, evidence: evidence}
}

func (s *Service) Apply(area string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.NewString()
	s.devices[id] = model.TelemetryEvidence{ID: id, AreaID: area, RecordedAt: time.Now().UTC()}
	return id
}

func (s *Service) Get(id string) (model.TelemetryEvidence, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.devices[id]
	return e, ok
}

func (s *Service) Records() []model.TelemetryEvidence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.TelemetryEvidence, 0, len(s.devices))
	for _, e := range s.devices {
		out = append(out, e)
	}
	return out
}

func (s *Service) Verify(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.devices[id]
	if !ok {
		return fmt.Errorf("ground not found")
	}
	if !s.collector.Zero(e.AreaID) {
		return fmt.Errorf("ground current remains")
	}
	e.ZeroCurrent = true
	e.Durable = true
	s.devices[id] = e
	return nil
}
