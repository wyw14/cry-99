package isolation

import (
	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Step struct {
	DeviceID  string             `json:"device_id"`
	Action    string             `json:"action"`
	CommandID string             `json:"command_id"`
	State     model.CommandState `json:"state"`
}

type Plan struct {
	ID        string    `json:"id"`
	PermitID  string    `json:"permit_id"`
	Steps     []Step    `json:"steps"`
	Current   int       `json:"current"`
	Cancelled bool      `json:"cancelled"`
	CreatedAt time.Time `json:"created_at"`
}

type Registry struct {
	mu    sync.RWMutex
	plans map[string]Plan
}

func NewRegistry() *Registry { return &Registry{plans: map[string]Plan{}} }

func (r *Registry) Create(permitID string, devices []string) Plan {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := Plan{ID: uuid.NewString(), PermitID: permitID, CreatedAt: time.Now().UTC()}
	for _, device := range devices {
		p.Steps = append(p.Steps, Step{DeviceID: device, Action: "open", State: model.CommandPlanned})
	}
	r.plans[p.ID] = p
	return p
}

func (r *Registry) Get(id string) (Plan, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plans[id]
	p.Steps = append([]Step(nil), p.Steps...)
	return p, ok
}

func (r *Registry) Save(p Plan) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.Steps = append([]Step(nil), p.Steps...)
	r.plans[p.ID] = p
}

func (r *Registry) List() []Plan {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Plan, 0, len(r.plans))
	for _, p := range r.plans {
		p.Steps = append([]Step(nil), p.Steps...)
		out = append(out, p)
	}
	return out
}
