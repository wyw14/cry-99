package earth

import (
	"example.com/railvolt/internal/model"
	"sync"
)

type Aggregate struct {
	mu       sync.RWMutex
	byArea   map[string]bool
	revision string
}

func NewAggregate() *Aggregate { return &Aggregate{byArea: map[string]bool{}} }

func (a *Aggregate) Rebuild(t model.Topology, records []model.TelemetryEvidence) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.byArea = map[string]bool{}
	a.revision = t.Revision
	for _, r := range records {
		if !r.Durable || !r.ZeroCurrent {
			a.byArea[r.AreaID] = true
		}
	}
}

func (a *Aggregate) Grounded(area string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.byArea[area]
}
func (a *Aggregate) Revision() string { a.mu.RLock(); defer a.mu.RUnlock(); return a.revision }
