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

func TestReadinessUsesCurrentSharedReturnTopology(t *testing.T) {
	top := topology.NewManager()
	current := top.Current()
	certificates := testcert.NewRegistry()
	certificates.Issue("A", current.Revision, current.Areas["A"].Boundaries, true, time.Hour)
	grounds := earth.NewAggregate()
	grounds.Rebuild(current, []model.TelemetryEvidence{{ID: "ground-B", AreaID: "B", Durable: false, ZeroCurrent: false}})
	eligibility := energize.NewEligibility(top, certificates, grounds, alarm.NewService())
	if err := eligibility.Ready("A", "permit", time.Now()); err == nil { t.Fatal("shared return ground was omitted from readiness") }
	grounds.Rebuild(current, nil)
	if err := eligibility.Ready("A", "permit", time.Now()); err != nil { t.Fatalf("readiness did not recover after neighboring ground removal: %v", err) }
}
