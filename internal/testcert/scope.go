package testcert

import (
	"fmt"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type Registry struct {
	mu    sync.RWMutex
	certs map[string]model.Certificate
}

func NewRegistry() *Registry { return &Registry{certs: make(map[string]model.Certificate)} }

func (r *Registry) Issue(areaID, revision string, scope []string, passed bool, ttl time.Duration) model.Certificate {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	c := model.Certificate{ID: uuid.NewString(), AreaID: areaID, Revision: revision, Scope: append([]string(nil), scope...), Passed: passed, IssuedAt: now, ExpiresAt: now.Add(ttl)}
	r.certs[c.ID] = c
	return c
}

func (r *Registry) Get(id string) (model.Certificate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.certs[id]
	return c, ok
}

func (r *Registry) ValidFor(areaID, revision string, now time.Time) (model.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var selected model.Certificate
	for _, c := range r.certs {
		if c.AreaID == areaID && c.Revision == revision && c.Current(now.Unix()) {
			selected = c
			break
		}
	}
	if selected.ID == "" {
		return model.Certificate{}, fmt.Errorf("no current certificate for area=%s revision=%s", areaID, revision)
	}
	return selected, nil
}
