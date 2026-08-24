package journal_test

import (
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
	"testing"
)

func TestCheckpointCommitsOneRevision(t *testing.T) {
	checkpoints := journal.NewCheckpoints()
	snapshot := model.Snapshot{Revision: 7, Permits: []model.Permit{{ID: "permit"}}}
	prepared, err := checkpoints.Prepare(snapshot, []journal.ComponentRevision{{Name: "permit", Revision: 7, Digest: "a"}, {Name: "switchgear", Revision: 7, Digest: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := checkpoints.Commit(prepared.ID); err != nil {
		t.Fatal(err)
	}
	recovered, err := checkpoints.Recover()
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Revision != 7 || !recovered.Committed {
		t.Fatalf("unexpected recovered snapshot: %+v", recovered)
	}
}

func TestCheckpointRejectsMixedRevision(t *testing.T) {
	checkpoints := journal.NewCheckpoints()
	prepared, err := checkpoints.Prepare(model.Snapshot{Revision: 8}, []journal.ComponentRevision{{Name: "permit", Revision: 8, Digest: "a"}, {Name: "evidence", Revision: 7, Digest: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := checkpoints.Commit(prepared.ID); err == nil {
		t.Fatal("expected mixed revision to be rejected")
	}
}
