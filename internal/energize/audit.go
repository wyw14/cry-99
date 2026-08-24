package energize

import (
	"fmt"
	"sort"
	"time"

	"example.com/railvolt/internal/model"
)

type StepAudit struct {
	Index       int                `json:"index"`
	CommandID   string             `json:"command_id"`
	DeviceID    string             `json:"device_id"`
	State       model.CommandState `json:"state"`
	Expected    string             `json:"expected"`
	Error       string             `json:"error,omitempty"`
	Recoverable bool               `json:"recoverable"`
}

type OperationAudit struct {
	PermitID       string      `json:"permit_id"`
	GeneratedAt    time.Time   `json:"generated_at"`
	CurrentStep    int         `json:"current_step"`
	TotalSteps     int         `json:"total_steps"`
	ConfirmedSteps int         `json:"confirmed_steps"`
	PendingSteps   int         `json:"pending_steps"`
	FailedSteps    int         `json:"failed_steps"`
	Steps          []StepAudit `json:"steps"`
	Consistent     bool        `json:"consistent"`
}

func AuditOperation(operation Operation) OperationAudit {
	report := OperationAudit{
		PermitID:    operation.PermitID,
		GeneratedAt: time.Now().UTC(),
		CurrentStep: operation.Next,
		TotalSteps:  len(operation.Commands),
		Consistent:  true,
	}
	for index, command := range operation.Commands {
		step := StepAudit{
			Index:       index,
			CommandID:   command.ID,
			DeviceID:    command.DeviceID,
			State:       command.State,
			Expected:    expectedState(index, operation.Next, operation.Complete),
			Error:       command.Error,
			Recoverable: command.State != model.CommandFailed,
		}
		switch command.State {
		case model.CommandConfirmed:
			report.ConfirmedSteps++
		case model.CommandFailed:
			report.FailedSteps++
		default:
			report.PendingSteps++
		}
		if index < operation.Next && command.State != model.CommandConfirmed {
			report.Consistent = false
			step.Recoverable = false
		}
		if index >= operation.Next && command.State == model.CommandConfirmed && !operation.Complete {
			report.Consistent = false
		}
		report.Steps = append(report.Steps, step)
	}
	if operation.Complete && report.ConfirmedSteps != report.TotalSteps {
		report.Consistent = false
	}
	if operation.Failed != "" && report.FailedSteps == 0 {
		report.Consistent = false
	}
	return report
}

func (a OperationAudit) Validate() error {
	if a.PermitID == "" {
		return fmt.Errorf("operation audit has no permit identity")
	}
	if a.TotalSteps != len(a.Steps) {
		return fmt.Errorf("operation audit step count differs")
	}
	if a.ConfirmedSteps+a.PendingSteps+a.FailedSteps != a.TotalSteps {
		return fmt.Errorf("operation audit counters do not cover every step")
	}
	if !a.Consistent {
		return fmt.Errorf("operation state and command boundary are inconsistent")
	}
	return nil
}

func (a OperationAudit) FailedDevices() []string {
	result := []string{}
	for _, step := range a.Steps {
		if step.State == model.CommandFailed {
			result = append(result, step.DeviceID)
		}
	}
	sort.Strings(result)
	return result
}

func (a OperationAudit) ResumeIndex() int {
	for _, step := range a.Steps {
		if step.State != model.CommandConfirmed {
			return step.Index
		}
	}
	return len(a.Steps)
}

func expectedState(index, next int, complete bool) string {
	if complete || index < next {
		return string(model.CommandConfirmed)
	}
	if index == next {
		return "next_or_in_flight"
	}
	return string(model.CommandPlanned)
}
