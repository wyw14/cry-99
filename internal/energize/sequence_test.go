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

func TestConfirmedDevicesAdvanceSequence(t *testing.T) {
	permits := permit.NewService()
	p := permits.Create("WS", "crew", "rev")
	p.Phase = model.PhaseReadyToEnergize
	p.EnergizeLocked = false
	if err := permits.Save(p); err != nil {
		t.Fatal(err)
	}
	queue := switchgear.NewQueue()
	client := switchgear.NewClient()
	alarms := alarm.NewService()
	sequence := energize.NewSequencer(queue, client, permits, alarms)
	sequence.Plan(p.ID, []string{"CB-1", "CB-2"})
	if _, err := sequence.Step(context.Background(), p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := sequence.Step(context.Background(), p.ID); err != nil {
		t.Fatal(err)
	}
	updated, _ := permits.Get(p.ID)
	if updated.Phase != model.PhaseEnergized || len(alarms.ForPermit(p.ID)) != 0 {
		t.Fatalf("confirmed sequence did not finish: permit=%+v alarms=%+v", updated, alarms.List())
	}
}

func TestOperationAuditMatchesBoundary(t *testing.T) {
	operation := energize.Operation{PermitID: "permit-1", Next: 1, Commands: []model.Command{{ID: "c1", DeviceID: "d1", State: model.CommandConfirmed}, {ID: "c2", DeviceID: "d2", State: model.CommandPlanned}}}
	report := energize.AuditOperation(operation)
	if err := report.Validate(); err != nil {
		t.Fatal(err)
	}
	if report.ResumeIndex() != 1 {
		t.Fatalf("unexpected resume boundary %d", report.ResumeIndex())
	}
}
