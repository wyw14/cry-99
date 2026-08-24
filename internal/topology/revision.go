package topology

import (
	"fmt"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type Manager struct {
	mu       sync.RWMutex
	current  model.Topology
	versions map[string]model.Topology
}

func NewManager() *Manager {
	initial := model.Topology{
		Revision: uuid.NewString(),
		Areas: map[string]model.Area{
			"A": {ID: "A", Name: "North siding", Feed: "F1", ReturnGroup: "R1", Boundaries: []string{"A-1", "A-2"}},
			"B": {ID: "B", Name: "South siding", Feed: "F1", ReturnGroup: "R1", Boundaries: []string{"B-1", "B-2"}},
		},
		SharedReturn: map[string][]string{"R1": {"A", "B"}},
		PublishedAt:  time.Now().UTC(),
	}
	return &Manager{current: initial, versions: map[string]model.Topology{initial.Revision: initial}}
}

func (m *Manager) Current() model.Topology {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneTopology(m.current)
}

func (m *Manager) Publish(t model.Topology) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.Revision == "" {
		t.Revision = uuid.NewString()
	}
	t.PublishedAt = time.Now().UTC()
	t = cloneTopology(t)
	m.current = t
	m.versions[t.Revision] = t
	return t.Revision
}

func (m *Manager) Revision() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.Revision
}

func (m *Manager) SplitArea(areaID string, first, second model.Area) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.current.Areas[areaID]; !ok {
		return "", fmt.Errorf("area %s not found", areaID)
	}
	next := cloneTopology(m.current)
	delete(next.Areas, areaID)
	next.Areas[first.ID] = first
	next.Areas[second.ID] = second
	next.Revision = uuid.NewString()
	next.PublishedAt = time.Now().UTC()
	m.current = next
	m.versions[next.Revision] = cloneTopology(next)
	return next.Revision, nil
}

func (m *Manager) Get(revision string) (model.Topology, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.versions[revision]
	if !ok {
		return model.Topology{}, false
	}
	return cloneTopology(t), true
}

func cloneTopology(t model.Topology) model.Topology {
	areas := make(map[string]model.Area, len(t.Areas))
	for k, v := range t.Areas {
		v.Boundaries = append([]string(nil), v.Boundaries...)
		areas[k] = v
	}
	returns := make(map[string][]string, len(t.SharedReturn))
	for k, v := range t.SharedReturn {
		returns[k] = append([]string(nil), v...)
	}
	t.Areas, t.SharedReturn = areas, returns
	return t
}
