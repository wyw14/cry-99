package energize

import (
	"example.com/railvolt/internal/topology"
	"time"
)

type Readiness struct {
	eligibility *Eligibility
	topology    *topology.Manager
}
type AreaReadiness struct {
	AreaID   string `json:"area_id"`
	Revision string `json:"revision"`
	Ready    bool   `json:"ready"`
	Reason   string `json:"reason,omitempty"`
}

func NewReadiness(e *Eligibility, t *topology.Manager) *Readiness {
	return &Readiness{eligibility: e, topology: t}
}

func (r *Readiness) View(permitID string) []AreaReadiness {
	t := r.topology.Current()
	out := make([]AreaReadiness, 0, len(t.Areas))
	for areaID := range t.Areas {
		item := AreaReadiness{AreaID: areaID, Revision: t.Revision}
		if err := r.eligibility.Ready(areaID, permitID, time.Now().UTC()); err != nil {
			item.Reason = err.Error()
		} else {
			item.Ready = true
		}
		out = append(out, item)
	}
	return out
}
