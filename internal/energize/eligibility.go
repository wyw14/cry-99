package energize

import (
	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/earth"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/testcert"
	"example.com/railvolt/internal/topology"
	"fmt"
	"time"
)

type Eligibility struct {
	topology     *topology.Manager
	certificates *testcert.Registry
	grounds      *earth.Aggregate
	alarms       *alarm.Service
}

func NewEligibility(t *topology.Manager, c *testcert.Registry, g *earth.Aggregate, a *alarm.Service) *Eligibility {
	return &Eligibility{topology: t, certificates: c, grounds: g, alarms: a}
}

func (e *Eligibility) Ready(areaID, permitID string, now time.Time) error {
	t := e.topology.Current()
	area, ok := t.Areas[areaID]
	if !ok {
		return fmt.Errorf("unknown area")
	}
	for _, shared := range e.topology.SharedAreas(areaID) {
		if e.grounds.Grounded(shared) {
			return fmt.Errorf("shared return path still grounded")
		}
	}
	cert, ok := e.certificates.ValidForArea(area, t.Revision, now)
	if !ok || !testcert.ScopeMatches(cert, area) {
		return fmt.Errorf("certificate does not cover current topology")
	}
	if alarm.BlocksEnergize(e.alarms.List(), permitID) {
		return fmt.Errorf("active alarm blocks energize")
	}
	return nil
}

func PermitReady(p model.Permit) bool {
	return p.Phase == model.PhaseReadyToEnergize && !p.EnergizeLocked && p.FailureReason == ""
}
