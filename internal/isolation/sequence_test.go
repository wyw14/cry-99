package isolation_test

import (
	"context"
	"testing"

	"example.com/railvolt/internal/isolation"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
)

func TestPlanTracksDeviceConfirmation(t *testing.T) {
	plans := isolation.NewRegistry()
	queue := switchgear.NewQueue()
	service := isolation.NewService(plans, queue)
	plan := service.Begin("permit-1", []string{"DS-1", "DS-2"})
	if len(plan.Steps) != 2 {
		t.Fatalf("expected two isolation steps")
	}
	if _, err := service.ExecuteNext(context.Background(), plan.ID); err != nil {
		t.Fatal(err)
	}
	confirmed, err := service.Confirm(plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Current != 1 || confirmed.Steps[0].State != model.CommandConfirmed {
		t.Fatalf("unexpected confirmation boundary: %+v", confirmed)
	}
}
