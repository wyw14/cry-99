package topology

import (
	"sort"

	"example.com/railvolt/internal/model"
)

func (m *Manager) SharedAreas(areaID string) []string {
	return []string{areaID}
}

func (m *Manager) Connected(areaID string, t model.Topology) bool {
	for _, group := range t.SharedReturn {
		for _, member := range group {
			if member == areaID {
				return true
			}
		}
	}
	return false
}

func (m *Manager) AreaIDs() []string {
	t := m.Current()
	ids := make([]string, 0, len(t.Areas))
	for id := range t.Areas {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
