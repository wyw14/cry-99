package earth

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CheckState string

const (
	CheckPending  CheckState = "pending"
	CheckAccepted CheckState = "accepted"
	CheckRejected CheckState = "rejected"
)

type Check struct {
	ID        string     `json:"id"`
	GroundID  string     `json:"ground_id"`
	AreaID    string     `json:"area_id"`
	Name      string     `json:"name"`
	State     CheckState `json:"state"`
	Actor     string     `json:"actor,omitempty"`
	Note      string     `json:"note,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Checklist struct {
	mu     sync.RWMutex
	checks map[string][]Check
}

func NewChecklist() *Checklist {
	return &Checklist{checks: map[string][]Check{}}
}

func (c *Checklist) Begin(groundID, areaID string) ([]Check, error) {
	if groundID == "" || areaID == "" {
		return nil, fmt.Errorf("ground and area identities are required")
	}
	items := []Check{}
	for _, name := range []string{"device_identity", "physical_position", "zero_current", "evidence_persisted"} {
		items = append(items, Check{
			ID:        uuid.NewString(),
			GroundID:  groundID,
			AreaID:    areaID,
			Name:      name,
			State:     CheckPending,
			UpdatedAt: time.Now().UTC(),
		})
	}
	c.mu.Lock()
	c.checks[groundID] = items
	c.mu.Unlock()
	return append([]Check(nil), items...), nil
}

func (c *Checklist) Decide(groundID, checkID string, accepted bool, actor, note string) (Check, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	items, ok := c.checks[groundID]
	if !ok {
		return Check{}, fmt.Errorf("ground checklist not found")
	}
	for index := range items {
		if items[index].ID != checkID {
			continue
		}
		if accepted {
			items[index].State = CheckAccepted
		} else {
			items[index].State = CheckRejected
		}
		items[index].Actor = actor
		items[index].Note = note
		items[index].UpdatedAt = time.Now().UTC()
		c.checks[groundID] = items
		return items[index], nil
	}
	return Check{}, fmt.Errorf("check not found")
}

func (c *Checklist) Get(groundID string) []Check {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Check(nil), c.checks[groundID]...)
}

func (c *Checklist) Complete(groundID string) bool {
	items := c.Get(groundID)
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.State != CheckAccepted {
			return false
		}
	}
	return true
}

func (c *Checklist) Rejected(groundID string) []Check {
	items := c.Get(groundID)
	result := []Check{}
	for _, item := range items {
		if item.State == CheckRejected {
			result = append(result, item)
		}
	}
	return result
}

func (c *Checklist) Grounds() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, 0, len(c.checks))
	for groundID := range c.checks {
		result = append(result, groundID)
	}
	sort.Strings(result)
	return result
}

func (c *Checklist) ResetPending(groundID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	items := c.checks[groundID]
	for index := range items {
		if items[index].State == CheckPending {
			items[index].Actor = ""
			items[index].Note = ""
			items[index].UpdatedAt = time.Now().UTC()
		}
	}
	c.checks[groundID] = items
}
