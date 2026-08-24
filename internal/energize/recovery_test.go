package energize_test

import (
	"testing"

	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
)

// TestRestoreOperationPreservesConfirmedBoundary guards the nightly-seal restart
// path: a power permit is mid-sequence across several breakers when the service
// restarts. Already-confirmed device actions (accepted by the field) must not be
// replayed; only the in-flight and still-planned commands should remain pending.
func TestRestoreOperationPreservesConfirmedBoundary(t *testing.T) {
	queue := switchgear.NewQueue()
	snapshot := model.Snapshot{
		Revision: 1,
		Commands: []model.Command{
			{ID: "cb-1", DeviceID: "CB-1", State: model.CommandConfirmed},
			{ID: "cb-2", DeviceID: "CB-2", State: model.CommandConfirmed},
			{ID: "cb-3", DeviceID: "CB-3", State: model.CommandAccepted},
			{ID: "cb-4", DeviceID: "CB-4", State: model.CommandPlanned},
		},
		NextCommand: 2,
		Committed:   true,
	}

	operation := energize.RestoreOperation(snapshot, queue)

	if operation.Next != 2 {
		t.Fatalf("restore must resume at the confirmed boundary: got Next=%d want 2", operation.Next)
	}
	if operation.Complete {
		t.Fatalf("operation must not be complete while in-flight commands remain")
	}

	resume := energize.ResumeCommands(snapshot)
	if len(resume) != 2 {
		t.Fatalf("only the in-flight and pending commands should be resumable: got %d want 2", len(resume))
	}
	if resume[0].ID != "cb-3" || resume[1].ID != "cb-4" {
		var ids []string
		for _, c := range resume {
			ids = append(ids, c.ID)
		}
		t.Fatalf("resumable commands must skip confirmed devices: got %v want [cb-3 cb-4]", ids)
	}

	// The restored queue must mark commands below the boundary as confirmed so
	// the resumed sequencer does not re-Begin/re-Execute the devices it already
	// accepted on the field before the restart.
	for i, cmd := range operation.Commands {
		if i < operation.Next && cmd.State != model.CommandConfirmed {
			t.Fatalf("command %s below boundary must remain confirmed: got %s", cmd.ID, cmd.State)
		}
	}
	if got := len(queue.Confirmed()); got != 2 {
		t.Fatalf("restored queue confirmed boundary must be 2: got %d", got)
	}
}
