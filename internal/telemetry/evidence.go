package telemetry

import (
	"example.com/railvolt/internal/model"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type EvidenceStore struct {
	mu         sync.RWMutex
	records    map[string]model.TelemetryEvidence
	failWrites bool
}

func NewEvidenceStore() *EvidenceStore {
	return &EvidenceStore{records: map[string]model.TelemetryEvidence{}}
}
func (s *EvidenceStore) Save(area string, samples []model.Sample, zero bool) (model.TelemetryEvidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failWrites {
		return model.TelemetryEvidence{}, fmt.Errorf("telemetry persistence unavailable")
	}
	e := model.TelemetryEvidence{ID: uuid.NewString(), AreaID: area, Samples: append([]model.Sample(nil), samples...), ZeroCurrent: zero, Durable: true, RecordedAt: time.Now().UTC()}
	s.records[e.ID] = e
	return e, nil
}
func (s *EvidenceStore) Get(id string) (model.TelemetryEvidence, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.records[id]
	return e, ok
}
func (s *EvidenceStore) All() []model.TelemetryEvidence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.TelemetryEvidence, 0, len(s.records))
	for _, e := range s.records {
		out = append(out, e)
	}
	return out
}
