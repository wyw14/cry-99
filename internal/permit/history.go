package permit

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type Transition struct {
	ID         string      `json:"id"`
	PermitID   string      `json:"permit_id"`
	Generation uint64      `json:"generation"`
	From       model.Phase `json:"from"`
	To         model.Phase `json:"to"`
	Actor      string      `json:"actor"`
	Reason     string      `json:"reason"`
	OccurredAt time.Time   `json:"occurred_at"`
}

type History struct {
	mu          sync.RWMutex
	transitions map[string][]Transition
}

func NewHistory() *History {
	return &History{transitions: map[string][]Transition{}}
}

func (h *History) Record(permit model.Permit, from, to model.Phase, actor, reason string) (Transition, error) {
	if permit.ID == "" {
		return Transition{}, fmt.Errorf("permit identity is required")
	}
	if from == to {
		return Transition{}, fmt.Errorf("transition must change phase")
	}
	entry := Transition{
		ID:         uuid.NewString(),
		PermitID:   permit.ID,
		Generation: permit.Generation,
		From:       from,
		To:         to,
		Actor:      actor,
		Reason:     reason,
		OccurredAt: time.Now().UTC(),
	}
	h.mu.Lock()
	h.transitions[permit.ID] = append(h.transitions[permit.ID], entry)
	h.mu.Unlock()
	return entry, nil
}

func (h *History) ForPermit(permitID string) []Transition {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]Transition(nil), h.transitions[permitID]...)
}

func (h *History) All() []Transition {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := []Transition{}
	for _, items := range h.transitions {
		result = append(result, items...)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].OccurredAt.Before(result[j].OccurredAt)
	})
	return result
}

func (h *History) Latest(permitID string) (Transition, bool) {
	items := h.ForPermit(permitID)
	if len(items) == 0 {
		return Transition{}, false
	}
	return items[len(items)-1], true
}

func (h *History) Validate(permit model.Permit) error {
	items := h.ForPermit(permit.ID)
	if len(items) == 0 {
		return nil
	}
	generation := items[0].Generation
	phase := items[0].From
	for _, item := range items {
		if item.Generation != generation {
			return fmt.Errorf("permit history crosses generation %d -> %d", generation, item.Generation)
		}
		if item.From != phase {
			return fmt.Errorf("permit history discontinuity %s -> %s", phase, item.From)
		}
		phase = item.To
	}
	if phase != permit.Phase {
		return fmt.Errorf("history phase %s differs from permit phase %s", phase, permit.Phase)
	}
	return nil
}

func (h *History) Generations() map[string]uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make(map[string]uint64, len(h.transitions))
	for permitID, items := range h.transitions {
		for _, item := range items {
			if item.Generation > result[permitID] {
				result[permitID] = item.Generation
			}
		}
	}
	return result
}
