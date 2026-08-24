package energize_test

import (
	"context"
	"testing"

	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
	"example.com/railvolt/internal/switchgear"
)

func TestSwitchRejectionPropagatesThroughEnergize(t *testing.T) {
	permits := permit.NewService()
	p := permits.Create("WS", "crew", "revision")
	p.Phase = model.PhaseReadyToEnergize
	p.EnergizeLocked = false
	permits.Save(p)
	queue := switchgear.NewQueue()
	client := switchgear.NewClient()
	client.Reject("CB-final", "remote lock active")
	alarms := alarm.NewService()
	sequence := energize.NewSequencer(queue, client, permits, alarms)
	sequence.Plan(p.ID, []string{"CB-final"})
	if _, err := sequence.Step(context.Background(), p.ID); err == nil { t.Fatal("controller rejection was reported as success") }
	updated, _ := permits.Get(p.ID)
	if updated.Phase == model.PhaseEnergized { t.Fatalf("rejected command energized permit: %+v", updated) }
	if len(alarms.ForPermit(p.ID)) != 1 { t.Fatalf("device rejection did not reach alarms: %+v", alarms.List()) }
	op, _ := sequence.Get(p.ID)
	if op.Failed == "" || op.Complete { t.Fatalf("sequence did not expose device rejection: %+v", op) }
}
