package journal_test

import (
	"testing"

	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
)

// TestSnapshotFromPersistsConfirmedBoundary guards the nightly-seal path: the
// number of consecutive confirmed device commands must be carried on the
// snapshot so a restart can resume at the boundary instead of replaying the
// already-accepted field actions.
func TestSnapshotFromPersistsConfirmedBoundary(t *testing.T) {
	commands := []model.Command{
		{ID: "cb-1", State: model.CommandConfirmed},
		{ID: "cb-2", State: model.CommandConfirmed},
		{ID: "cb-3", State: model.CommandAccepted},
		{ID: "cb-4", State: model.CommandPlanned},
	}
	snapshot := journal.SnapshotFrom(nil, commands, model.Topology{}, nil, 2, 9)
	if !snapshot.Committed {
		t.Fatalf("checkpoint snapshot must be committed")
	}
	if snapshot.NextCommand != 2 {
		t.Fatalf("snapshot must carry confirmed boundary: got %d want 2", snapshot.NextCommand)
	}
	pending := journal.PendingCommands(snapshot)
	if len(pending) != 2 {
		t.Fatalf("pending commands must exclude confirmed boundary: got %d want 2", len(pending))
	}
}

// TestSnapshotFromClampsBoundary guarantees the persisted boundary never escapes
// the command list, so a corrupt or off-by-one count cannot pull the resume
// pointer past the sequence on recovery.
func TestSnapshotFromClampsBoundary(t *testing.T) {
	commands := []model.Command{
		{ID: "cb-1", State: model.CommandConfirmed},
		{ID: "cb-2", State: model.CommandConfirmed},
	}
	for _, next := range []int{-3, 42} {
		snapshot := journal.SnapshotFrom(nil, commands, model.Topology{}, nil, next, 9)
		if snapshot.NextCommand < 0 || snapshot.NextCommand > len(commands) {
			t.Fatalf("boundary %d not clamped: got %d", next, snapshot.NextCommand)
		}
	}
}
