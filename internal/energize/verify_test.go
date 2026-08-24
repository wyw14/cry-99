package energize_test

import (
	"context"
	"testing"

	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/permit"
	"example.com/railvolt/internal/switchgear"
)

func TestTimedOutRequestStopsFurtherEnergizeCommands(t *testing.T) {
	permits := permit.NewService()
	p := permits.Create("WS", "crew", "revision")
	queue := switchgear.NewQueue()
	client := switchgear.NewClient()
	sequence := energize.NewSequencer(queue, client, permits, alarm.NewService())
	sequence.Plan(p.ID, []string{"CB-1", "CB-2", "CB-3"})
	runner := energize.NewRunner(sequence)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := <-runner.Run(ctx, p.ID); err == nil { t.Fatal("cancelled request reported success") }
	if calls := client.Calls(); len(calls) != 0 { t.Fatalf("commands continued after request cancellation: %+v", calls) }
	op, _ := sequence.Get(p.ID)
	if op.Next != 0 || op.Failed == "" { t.Fatalf("cancelled boundary was not retained: %+v", op) }
}
