package crew

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Member struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	Qualification string    `json:"qualification"`
	ValidUntil    time.Time `json:"valid_until"`
}

type Shift struct {
	ID        string    `json:"id"`
	Worksite  string    `json:"worksite"`
	Owner     string    `json:"owner"`
	Members   []Member  `json:"members"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
}

type Roster struct {
	mu      sync.RWMutex
	shifts  map[string]Shift
	current map[string]string
}

func NewRoster() *Roster {
	return &Roster{shifts: map[string]Shift{}, current: map[string]string{}}
}

func (r *Roster) Start(worksite, owner string, members []Member) (Shift, error) {
	if worksite == "" || owner == "" {
		return Shift{}, fmt.Errorf("worksite and owner are required")
	}
	if len(members) == 0 {
		return Shift{}, fmt.Errorf("at least one qualified member is required")
	}
	now := time.Now().UTC()
	for index := range members {
		if members[index].ID == "" {
			members[index].ID = uuid.NewString()
		}
		if members[index].ValidUntil.Before(now) {
			return Shift{}, fmt.Errorf("qualification expired for %s", members[index].Name)
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if activeID := r.current[worksite]; activeID != "" {
		active := r.shifts[activeID]
		if active.EndedAt.IsZero() {
			return Shift{}, fmt.Errorf("worksite already has active shift")
		}
	}
	shift := Shift{ID: uuid.NewString(), Worksite: worksite, Owner: owner, Members: append([]Member(nil), members...), StartedAt: now}
	r.shifts[shift.ID] = shift
	r.current[worksite] = shift.ID
	return shift, nil
}

func (r *Roster) End(id string) (Shift, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	shift, ok := r.shifts[id]
	if !ok {
		return Shift{}, fmt.Errorf("shift not found")
	}
	if !shift.EndedAt.IsZero() {
		return shift, nil
	}
	shift.EndedAt = time.Now().UTC()
	r.shifts[id] = shift
	if r.current[shift.Worksite] == id {
		delete(r.current, shift.Worksite)
	}
	return shift, nil
}

func (r *Roster) Current(worksite string) (Shift, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id := r.current[worksite]
	shift, ok := r.shifts[id]
	shift.Members = append([]Member(nil), shift.Members...)
	return shift, ok
}

func (r *Roster) List() []Shift {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Shift, 0, len(r.shifts))
	for _, shift := range r.shifts {
		shift.Members = append([]Member(nil), shift.Members...)
		result = append(result, shift)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartedAt.Before(result[j].StartedAt)
	})
	return result
}

func (r *Roster) Qualified(worksite, role string, at time.Time) bool {
	shift, ok := r.Current(worksite)
	if !ok || !shift.EndedAt.IsZero() {
		return false
	}
	for _, member := range shift.Members {
		if member.Role == role && member.ValidUntil.After(at) {
			return true
		}
	}
	return false
}
