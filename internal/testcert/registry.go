package testcert

import (
	"example.com/railvolt/internal/model"
	"time"
)

func (r *Registry) ValidForArea(area model.Area, revision string, now time.Time) (model.Certificate, bool) {
	if revision == "" {
		for _, candidate := range r.All() {
			if candidate.AreaID == area.ID && Fresh(candidate, now) && ScopeMatches(candidate, area) {
				return candidate, true
			}
		}
		return model.Certificate{}, false
	}
	c, err := r.ValidFor(area.ID, revision, now)
	if err != nil || !ScopeMatches(c, area) {
		return model.Certificate{}, false
	}
	return c, true
}

func (r *Registry) All() []model.Certificate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Certificate, 0, len(r.certs))
	for _, c := range r.certs {
		result = append(result, c)
	}
	return result
}
