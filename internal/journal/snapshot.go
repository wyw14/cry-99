package journal

import "example.com/railvolt/internal/model"

func SnapshotFrom(permits []model.Permit, commands []model.Command, topology model.Topology, evidence []model.TelemetryEvidence, next int, revision uint64) model.Snapshot {
	return model.Snapshot{Revision: revision, Permits: append([]model.Permit(nil), permits...), Commands: append([]model.Command(nil), commands...), Topology: topology, Evidence: append([]model.TelemetryEvidence(nil), evidence...), NextCommand: 0, Committed: true}
}
