package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type SafetyIssue struct {
	Component string    `json:"component"`
	Subject   string    `json:"subject"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Observed  time.Time `json:"observed"`
}

type SafetyReport struct {
	Revision       string        `json:"revision"`
	GeneratedAt    time.Time     `json:"generated_at"`
	PermitCount    int           `json:"permit_count"`
	CommandCount   int           `json:"command_count"`
	EvidenceCount  int           `json:"evidence_count"`
	AlarmCount     int           `json:"alarm_count"`
	Issues         []SafetyIssue `json:"issues"`
	ReadyPermits   []string      `json:"ready_permits"`
	BlockedPermits []string      `json:"blocked_permits"`
}

type AuditInput struct {
	Topology Topology
	Permits  []Permit
	Commands []Command
	Evidence []TelemetryEvidence
	Alarms   []Alarm
}

func AuditSafety(input AuditInput) SafetyReport {
	report := SafetyReport{
		Revision:      input.Topology.Revision,
		GeneratedAt:   time.Now().UTC(),
		PermitCount:   len(input.Permits),
		CommandCount:  len(input.Commands),
		EvidenceCount: len(input.Evidence),
		AlarmCount:    len(input.Alarms),
	}
	commandByPermit := make(map[string][]Command)
	for _, command := range input.Commands {
		commandByPermit[command.PermitID] = append(commandByPermit[command.PermitID], command)
	}
	alarmByPermit := make(map[string][]Alarm)
	for _, item := range input.Alarms {
		alarmByPermit[item.PermitID] = append(alarmByPermit[item.PermitID], item)
	}
	for _, permit := range input.Permits {
		report.inspectPermit(input.Topology, permit, commandByPermit[permit.ID], alarmByPermit[permit.ID])
	}
	for _, evidence := range input.Evidence {
		if !evidence.Durable {
			report.Issues = append(report.Issues, SafetyIssue{
				Component: "telemetry",
				Subject:   evidence.ID,
				Severity:  "error",
				Message:   "telemetry evidence is visible before durable persistence",
				Observed:  report.GeneratedAt,
			})
		}
		if evidence.Durable && len(evidence.Samples) == 0 {
			report.Issues = append(report.Issues, SafetyIssue{
				Component: "telemetry",
				Subject:   evidence.ID,
				Severity:  "warning",
				Message:   "durable evidence contains no electrical sample",
				Observed:  report.GeneratedAt,
			})
		}
	}
	sort.Strings(report.ReadyPermits)
	sort.Strings(report.BlockedPermits)
	sort.Slice(report.Issues, func(i, j int) bool {
		left := report.Issues[i].Severity + report.Issues[i].Component + report.Issues[i].Subject
		right := report.Issues[j].Severity + report.Issues[j].Component + report.Issues[j].Subject
		return left < right
	})
	return report
}

func (r *SafetyReport) inspectPermit(topology Topology, permit Permit, commands []Command, alarms []Alarm) {
	blocked := permit.EnergizeLocked || permit.GroundLocked || permit.FailureReason != ""
	if permit.TopologyRevision != topology.Revision {
		blocked = true
		r.Issues = append(r.Issues, SafetyIssue{
			Component: "permit",
			Subject:   permit.ID,
			Severity:  "warning",
			Message:   fmt.Sprintf("permit topology %s differs from active %s", permit.TopologyRevision, topology.Revision),
			Observed:  r.GeneratedAt,
		})
	}
	for _, command := range commands {
		if command.State == CommandFailed {
			blocked = true
			r.Issues = append(r.Issues, SafetyIssue{
				Component: "switchgear",
				Subject:   command.ID,
				Severity:  "error",
				Message:   strings.TrimSpace(command.Error),
				Observed:  r.GeneratedAt,
			})
		}
		if permit.Phase == PhaseEnergized && command.State != CommandConfirmed {
			blocked = true
			r.Issues = append(r.Issues, SafetyIssue{
				Component: "energize",
				Subject:   permit.ID,
				Severity:  "error",
				Message:   "energized permit contains an unconfirmed device command",
				Observed:  r.GeneratedAt,
			})
		}
	}
	if len(alarms) > 0 {
		blocked = true
	}
	if blocked {
		r.BlockedPermits = append(r.BlockedPermits, permit.ID)
	} else {
		r.ReadyPermits = append(r.ReadyPermits, permit.ID)
	}
}

func (r SafetyReport) Healthy() bool {
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			return false
		}
	}
	return true
}

func (r SafetyReport) IssueCounts() map[string]int {
	counts := map[string]int{"error": 0, "warning": 0, "info": 0}
	for _, issue := range r.Issues {
		counts[issue.Severity]++
	}
	return counts
}
