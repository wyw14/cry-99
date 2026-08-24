package api

import (
	"fmt"
	"net/http"

	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/permit"
)

func (a *Application) createCheckpoint(w http.ResponseWriter, r *http.Request) {
	revision := a.Journal.Revision() + 1
	next := 0
	commands := a.Commands.All()
	for next < len(commands) && commands[next].Complete() {
		next++
	}
	snapshot := journal.SnapshotFrom(a.Permits.Snapshot(), commands, a.Topology.Current(), a.TelemetryLog.All(), next, revision)
	components := []journal.ComponentRevision{
		{Name: "permit", Revision: revision, Digest: fmt.Sprintf("permits-%d", len(snapshot.Permits))},
		{Name: "switchgear", Revision: revision, Digest: fmt.Sprintf("commands-%d", len(snapshot.Commands))},
		{Name: "topology", Revision: revision, Digest: snapshot.Topology.Revision},
		{Name: "telemetry", Revision: revision, Digest: fmt.Sprintf("evidence-%d", len(snapshot.Evidence))},
	}
	prepared, err := a.Checkpoints.Prepare(snapshot, components)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	committed, err := a.Checkpoints.Commit(prepared.ID)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	a.Journal.SaveSnapshot(committed.Snapshot)
	a.Sealer.Seal(committed.Snapshot)
	writeJSON(w, http.StatusCreated, committed)
}

func (a *Application) recoveryStatus(w http.ResponseWriter, r *http.Request) {
	snapshot, err := a.Checkpoints.Recover()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "revisions": a.Checkpoints.CompleteRevisions()})
		return
	}
	restoredPermits := permit.Restore(snapshot.Permits)
	operation := energize.RestoreOperation(snapshot, a.Commands)
	writeJSON(w, http.StatusOK, map[string]any{
		"available":        true,
		"revision":         snapshot.Revision,
		"permit_count":     len(restoredPermits.List()),
		"pending_commands": energize.ResumeCommands(snapshot),
		"operation_audit":  energize.AuditOperation(operation),
	})
}
