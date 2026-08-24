package journal_test

import (
	"testing"

	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
	"example.com/railvolt/internal/switchgear"
)

func TestRecoveryUsesOneCommittedRevision(t *testing.T) {
	permitSnapshot := model.Snapshot{Revision: 12, Committed: true, Permits: []model.Permit{{ID: "permit", Worksite: "WS", Phase: model.PhaseReadyToEnergize, Generation: 3}}}
	evidenceSnapshot := model.Snapshot{Revision: 11, Committed: true, Commands: []model.Command{{ID: "command", DeviceID: "CB", Epoch: 11, State: model.CommandMoving}}}
	if _, err := journal.RecoverSingleRevision(permitSnapshot, evidenceSnapshot); err == nil { t.Fatal("recovery combined different component revisions") }
	matching := evidenceSnapshot
	matching.Revision = 12
	combined, err := journal.RecoverSingleRevision(permitSnapshot, matching)
	if err != nil { t.Fatal(err) }
	restored := permit.Restore(combined.Permits)
	p, ok := restored.Get("permit")
	if !ok || p.Generation != 3 || p.Phase != model.PhaseReadyToEnergize { t.Fatalf("permit snapshot changed during restore: %+v", p) }
	store := switchgear.NewEvidenceStore()
	for _, command := range matching.Commands { store.Put(command) }
	if items := store.All(); len(items) != 1 || items[0].Epoch != 11 { t.Fatalf("device evidence revision changed during restore: %+v", items) }
}
