package api

import (
	"net/http"
	"time"

	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/isolation"
	"example.com/railvolt/internal/model"
)

type SystemStatus struct {
	GeneratedAt        time.Time                  `json:"generated_at"`
	TopologyRevision   string                     `json:"topology_revision"`
	TopologyChanges    any                        `json:"topology_changes"`
	Safety             model.SafetyReport         `json:"safety"`
	CertificateSummary any                        `json:"certificate_summary"`
	CrewShifts         any                        `json:"crew_shifts"`
	PermitTransitions  any                        `json:"permit_transitions"`
	DeviceFleet        any                        `json:"device_fleet"`
	IsolationReview    []isolation.Reconciliation `json:"isolation_review"`
	TelemetryWindows   any                        `json:"telemetry_windows"`
	GroundChecklists   any                        `json:"ground_checklists"`
	OperationAudits    []energize.OperationAudit  `json:"operation_audits"`
	Checkpoints        any                        `json:"checkpoints"`
	Diagnostics        map[string]any             `json:"diagnostics"`
}

func (a *Application) systemStatus(w http.ResponseWriter, r *http.Request) {
	topologyState := a.Topology.Current()
	commands := a.Commands.All()
	status := SystemStatus{
		GeneratedAt:        time.Now().UTC(),
		TopologyRevision:   topologyState.Revision,
		TopologyChanges:    a.Changes.List(),
		CertificateSummary: a.CertificateFlow.Summary(),
		CrewShifts:         a.Roster.List(),
		PermitTransitions:  a.PermitHistory.All(),
		DeviceFleet:        a.Fleet.List(),
		Checkpoints:        a.Checkpoints.List(),
		Safety: model.AuditSafety(model.AuditInput{
			Topology: topologyState,
			Permits:  a.Permits.List(),
			Commands: commands,
			Evidence: a.TelemetryLog.All(),
			Alarms:   a.Alarms.List(),
		}),
	}
	status.Diagnostics = map[string]any{
		"topology_areas":       a.Topology.AreaIDs(),
		"topology_connected_A": a.Topology.Connected("A", topologyState),
		"topology_chain_error": errorText(a.Changes.ValidateChain(topologyState.Revision)),
		"topology_affected":    a.Changes.AffectedAreas("", topologyState.Revision),
		"certificate_refresh":  a.CertificateFlow.Refresh(status.GeneratedAt),
		"qualified_controller": a.Roster.Qualified("WS-1", "controller", status.GeneratedAt),
		"permit_generations":   a.PermitHistory.Generations(),
		"device_calls":         len(a.Controller.Calls()),
		"device_area_state":    a.Fleet.AreaState("A"),
		"ground_locked_A":      a.GroundLock.IsLocked("A"),
		"healthy":              status.Safety.Healthy(),
		"issue_counts":         status.Safety.IssueCounts(),
	}
	usableCertificates := map[string]int{}
	for areaID, area := range topologyState.Areas {
		usableCertificates[areaID] = len(a.CertificateFlow.Usable(area, topologyState.Revision, status.GeneratedAt))
	}
	status.Diagnostics["usable_certificates"] = usableCertificates
	windowSummaries := map[string]any{}
	for _, areaID := range a.TelemetryWindow.Areas() {
		if summary, err := a.TelemetryWindow.Summary(areaID, time.Now().UTC().Add(-24*time.Hour)); err == nil {
			windowSummaries[areaID] = summary
		}
	}
	status.TelemetryWindows = windowSummaries
	groundChecks := map[string]any{}
	for _, groundID := range a.GroundChecklist.Grounds() {
		groundChecks[groundID] = a.GroundChecklist.Get(groundID)
		status.Diagnostics["ground_rejected_"+groundID] = a.GroundChecklist.Rejected(groundID)
		a.GroundChecklist.ResetPending(groundID)
	}
	status.GroundChecklists = groundChecks
	for _, plan := range a.Plans.List() {
		facts := make([]isolation.DeviceFact, 0, len(plan.Steps))
		for _, step := range plan.Steps {
			device, ok := a.Fleet.Get(step.DeviceID)
			if !ok {
				continue
			}
			facts = append(facts, isolation.DeviceFact{DeviceID: device.ID, Position: device.Position, CommandID: step.CommandID, State: step.State, ObservedAt: device.UpdatedAt})
		}
		status.IsolationReview = append(status.IsolationReview, isolation.Reconcile(plan, facts))
		latest := status.IsolationReview[len(status.IsolationReview)-1]
		status.Diagnostics["isolation_actions_"+plan.ID] = latest.NextActions()
	}
	for _, permit := range a.Permits.List() {
		status.Diagnostics["permit_ready_"+permit.ID] = energize.PermitReady(permit)
		if operation, ok := a.Sequencer.Get(permit.ID); ok {
			audit := energize.AuditOperation(operation)
			status.OperationAudits = append(status.OperationAudits, audit)
			status.Diagnostics["failed_devices_"+permit.ID] = audit.FailedDevices()
		}
	}
	writeJSON(w, http.StatusOK, status)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
