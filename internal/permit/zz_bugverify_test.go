package permit_test

import (
	"testing"

	"example.com/railvolt/internal/crew"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
)

func TestStaleReceiptCannotAdvanceNewShiftPermit(t *testing.T) {
	permits := permit.NewService()
	mgr := crew.NewManager()

	// Day shift owned permit A, now in working phase.
	dayPermit := permits.Create("WS-1", "day", "rev")
	dayPermit.Phase = model.PhaseWorking
	permits.Save(dayPermit)
	daySession := mgr.Handover("WS-1", "day", 1)

	// Handover: night shift takes the same worksite.
	mgr.Handover("WS-1", "night", 1)
	// Night creates a new permit B at the same worksite (also generation 1).
	nightPermit := permits.Create("WS-1", "night", "rev")
	nightPermit.Phase = model.PhaseWorking
	permits.Save(nightPermit)

	// Day's stale terminal resends its release receipt for permit A.
	staleReceipt := model.Receipt{
		PermitID:   dayPermit.ID,
		Worksite:   "WS-1",
		Owner:      "day",
		Generation: 1,
		Fencing:    daySession.Fencing,
	}
	if _, err := permits.AcceptRelease(mgr, staleReceipt); err == nil {
		t.Fatal("stale day receipt must not advance night's permit")
	}

	// Night's own release receipt still works.
	nightSession, _ := mgr.Current("WS-1")
	nightReceipt := model.Receipt{
		PermitID:   nightPermit.ID,
		Worksite:   "WS-1",
		Owner:      nightSession.Owner,
		Generation: nightSession.Generation,
		Fencing:    nightSession.Fencing,
	}
	released, err := permits.AcceptRelease(mgr, nightReceipt)
	if err != nil {
		t.Fatalf("night release failed: %v", err)
	}
	if released.Phase != model.PhaseReadyToEnergize {
		t.Fatalf("night permit not advanced: %s", released.Phase)
	}
}
