package isolation

import (
	"fmt"
	"sort"
	"time"

	"example.com/railvolt/internal/model"
)

type DeviceFact struct {
	DeviceID   string             `json:"device_id"`
	Position   string             `json:"position"`
	CommandID  string             `json:"command_id"`
	State      model.CommandState `json:"state"`
	ObservedAt time.Time          `json:"observed_at"`
}

type Difference struct {
	DeviceID string `json:"device_id"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Action   string `json:"action"`
}

type Reconciliation struct {
	PlanID      string       `json:"plan_id"`
	PermitID    string       `json:"permit_id"`
	ComparedAt  time.Time    `json:"compared_at"`
	Differences []Difference `json:"differences"`
	Confirmed   []string     `json:"confirmed"`
	InFlight    []string     `json:"in_flight"`
	Pending     []string     `json:"pending"`
	Complete    bool         `json:"complete"`
}

func Reconcile(plan Plan, facts []DeviceFact) Reconciliation {
	result := Reconciliation{
		PlanID:     plan.ID,
		PermitID:   plan.PermitID,
		ComparedAt: time.Now().UTC(),
	}
	byDevice := make(map[string]DeviceFact, len(facts))
	for _, fact := range facts {
		previous, exists := byDevice[fact.DeviceID]
		if !exists || fact.ObservedAt.After(previous.ObservedAt) {
			byDevice[fact.DeviceID] = fact
		}
	}
	for _, step := range plan.Steps {
		fact, exists := byDevice[step.DeviceID]
		if !exists {
			result.Pending = append(result.Pending, step.DeviceID)
			result.Differences = append(result.Differences, Difference{
				DeviceID: step.DeviceID,
				Expected: expectedPosition(step.Action),
				Actual:   "unknown",
				Action:   "observe_device",
			})
			continue
		}
		expected := expectedPosition(step.Action)
		switch {
		case fact.State == model.CommandConfirmed && fact.Position == expected:
			result.Confirmed = append(result.Confirmed, step.DeviceID)
		case fact.State == model.CommandMoving || fact.State == model.CommandAccepted:
			result.InFlight = append(result.InFlight, step.DeviceID)
		case fact.State == model.CommandFailed:
			result.Differences = append(result.Differences, Difference{
				DeviceID: step.DeviceID,
				Expected: expected,
				Actual:   fact.Position,
				Action:   "investigate_failure",
			})
		default:
			result.Pending = append(result.Pending, step.DeviceID)
			if fact.Position != expected {
				result.Differences = append(result.Differences, Difference{
					DeviceID: step.DeviceID,
					Expected: expected,
					Actual:   fact.Position,
					Action:   "restore_plan_boundary",
				})
			}
		}
	}
	sort.Strings(result.Confirmed)
	sort.Strings(result.InFlight)
	sort.Strings(result.Pending)
	result.Complete = len(result.Confirmed) == len(plan.Steps) && len(result.Differences) == 0
	return result
}

func (r Reconciliation) Recoverable() bool {
	for _, difference := range r.Differences {
		if difference.Action == "investigate_failure" {
			return false
		}
	}
	return len(r.InFlight) > 0 || len(r.Pending) > 0 || r.Complete
}

func (r Reconciliation) NextActions() []string {
	actions := make([]string, 0, len(r.Differences)+len(r.InFlight))
	seen := map[string]struct{}{}
	for _, difference := range r.Differences {
		if _, exists := seen[difference.Action]; exists {
			continue
		}
		seen[difference.Action] = struct{}{}
		actions = append(actions, difference.Action)
	}
	if len(r.InFlight) > 0 {
		actions = append(actions, "wait_for_device_confirmation")
	}
	if len(actions) == 0 && r.Complete {
		actions = append(actions, "advance_permit")
	}
	return actions
}

func (r Reconciliation) Validate() error {
	if r.PlanID == "" || r.PermitID == "" {
		return fmt.Errorf("reconciliation identity is incomplete")
	}
	known := map[string]struct{}{}
	for _, device := range append(append(append([]string{}, r.Confirmed...), r.InFlight...), r.Pending...) {
		if _, exists := known[device]; exists {
			return fmt.Errorf("device %s appears in multiple reconciliation groups", device)
		}
		known[device] = struct{}{}
	}
	return nil
}

func expectedPosition(action string) string {
	switch action {
	case "open":
		return "open"
	case "close":
		return "closed"
	default:
		return "unknown"
	}
}
