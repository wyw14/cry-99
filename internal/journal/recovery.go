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
	if permitSnapshot.Revision != evidenceSnapshot.Revision || !permitSnapshot.Committed || !evidenceSnapshot.Committed {
		return model.Snapshot{}, fmt.Errorf("recovery revision split permit=%d evidence=%d", permitSnapshot.Revision, evidenceSnapshot.Revision)
	}
	return permitSnapshot, nil
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
