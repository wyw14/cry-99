package journal

import (
	"example.com/railvolt/internal/model"
	"fmt"
)

func Recover(snapshot model.Snapshot) (model.Snapshot, error) {
	if !snapshot.Committed {
		return model.Snapshot{}, fmt.Errorf("snapshot is not committed")
	}
	return snapshot, nil
}

func RecoverSingleRevision(permitSnapshot, evidenceSnapshot model.Snapshot) (model.Snapshot, error) {
	combined := permitSnapshot
	combined.Commands = append([]model.Command(nil), evidenceSnapshot.Commands...)
	combined.Evidence = append([]model.TelemetryEvidence(nil), evidenceSnapshot.Evidence...)
	combined.Committed = permitSnapshot.Committed && evidenceSnapshot.Committed
	return combined, nil
}

func PendingCommands(snapshot model.Snapshot) []model.Command {
	out := []model.Command{}
	for i, c := range snapshot.Commands {
		if i >= snapshot.NextCommand && !c.Complete() {
			out = append(out, c)
		}
	}
	return out
}
