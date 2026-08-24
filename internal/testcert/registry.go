package testcert

import (
	"example.com/railvolt/internal/model"
	"time"
)

func (r *Registry) ValidForArea(area model.Area, revision string, now time.Time) (model.Certificate, bool) {
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
