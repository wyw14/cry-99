package testcert

import (
	"example.com/railvolt/internal/model"
	"time"
)

func (r *Registry) ValidForArea(area model.Area, revision string, now time.Time) (model.Certificate, bool) {
	if revision == "" {
		for _, candidate := range r.All() {
			if candidate.AreaID == area.ID && Fresh(candidate, now) && containsScope(candidate.Scope, area.Boundaries) {
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

func containsScope(have, need []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, item := range have {
		set[item] = struct{}{}
	}
	for _, item := range need {
		if _, ok := set[item]; !ok {
			return false
		}
	}
	return true
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
