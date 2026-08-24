package energize

import (
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
)

func RestoreOperation(snapshot model.Snapshot, queue *switchgear.Queue) Operation {
	next := snapshot.NextCommand
	if next < 0 {
		next = 0
	}
	if next > len(snapshot.Commands) {
		next = len(snapshot.Commands)
	}
	queue.RestoreBoundary(snapshot.Commands, next)
	pending := journal.PendingCommands(snapshot)
	return Operation{Commands: append([]model.Command(nil), snapshot.Commands...), Next: next, Complete: len(pending) == 0}
}

func ResumeCommands(snapshot model.Snapshot) []model.Command {
	return journal.PendingCommands(snapshot)
}
