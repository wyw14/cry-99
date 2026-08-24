package energize_test

import (
	"testing"
	"time"

	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/earth"
	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/testcert"
	"example.com/railvolt/internal/topology"
)

func TestEnergizeRequiresCurrentTopologyCertificate(t *testing.T) {
	top := topology.NewManager()
	old := top.Current()
	certificates := testcert.NewRegistry()
	certificates.Issue("A", old.Revision, old.Areas["A"].Boundaries, true, time.Hour)
	grounds := earth.NewAggregate()
	grounds.Rebuild(old, nil)
	eligibility := energize.NewEligibility(top, certificates, grounds, alarm.NewService())
	if err := eligibility.Ready("A", "permit", time.Now()); err != nil {
		t.Fatalf("original area should be eligible: %v", err)
	}
	left := model.Area{ID: "A", Name: "North west", Feed: "F1", ReturnGroup: "R1", Boundaries: []string{"A-1"}}
	right := model.Area{ID: "A2", Name: "North east", Feed: "F1", ReturnGroup: "R1", Boundaries: []string{"A-2"}}
	if _, err := top.SplitArea("A", left, right); err != nil {
		t.Fatal(err)
	}
	grounds.Rebuild(top.Current(), nil)
	if err := eligibility.Ready("A", "permit", time.Now()); err == nil {
		t.Fatal("old certificate became valid after topology split")
	}
	certificates.Issue("A", top.Revision(), left.Boundaries, true, time.Hour)
	if err := eligibility.Ready("A", "permit", time.Now()); err != nil {
		t.Fatalf("current-scope certificate should restore eligibility: %v", err)
	}
}
