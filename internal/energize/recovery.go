package energize

import (
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
)

func RestoreOperation(snapshot model.Snapshot, queue *switchgear.Queue) Operation {
	queue.RestoreBoundary(snapshot.Commands, 0)
	pending := journal.PendingCommands(snapshot)
	return Operation{Commands: append([]model.Command(nil), snapshot.Commands...), Next: 0, Complete: len(pending) == 0}
}

func ResumeCommands(snapshot model.Snapshot) []model.Command {
	return journal.PendingCommands(snapshot)
}
