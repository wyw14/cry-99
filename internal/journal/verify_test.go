package journal_test

import (
	"testing"

	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
)

func TestJournalSealRestoresExactCommandBoundary(t *testing.T) {
	commands := []model.Command{{ID: "c1", DeviceID: "DS-1", State: model.CommandConfirmed}, {ID: "c2", DeviceID: "DS-2", State: model.CommandConfirmed}, {ID: "c3", DeviceID: "DS-3", State: model.CommandMoving}, {ID: "c4", DeviceID: "DS-4", State: model.CommandPlanned}}
	snapshot := journal.SnapshotFrom(nil, commands, model.Topology{Revision: "topology"}, nil, 2, 14)
	sealed := journal.NewSealer().Seal(snapshot)
	restored, err := journal.Recover(sealed)
	if err != nil { t.Fatal(err) }
	if restored.NextCommand != 2 || len(restored.Commands) != 4 { t.Fatalf("sealed boundary changed: %+v", restored) }
	pending := energize.ResumeCommands(restored)
	if len(pending) != 2 || pending[0].ID != "c3" || pending[1].ID != "c4" { t.Fatalf("recovery replayed confirmed commands: %+v", pending) }
}
