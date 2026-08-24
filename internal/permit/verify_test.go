package permit_test

import (
	"testing"

	"example.com/railvolt/internal/crew"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/permit"
)

func TestOldCrewReceiptCannotReleaseNewPermit(t *testing.T) {
	crews := crew.NewManager()
	oldSession := crews.Handover("WS-1", "day", 1)
	permits := permit.NewService()
	oldPermit := permits.Create("WS-1", "day", "r1")
	oldPermit.Phase = model.PhaseWorking
	permits.Save(oldPermit)
	crews.Handover("WS-1", "night", 2)
	newPermit := permits.Create("WS-1", "night", "r2")
	newPermit.Generation = 2
	newPermit.Phase = model.PhaseWorking
	permits.Save(newPermit)
	late := crew.HandoverReceipt(oldSession, oldPermit)
	late.PermitID = newPermit.ID
	if _, err := permits.AcceptRelease(crews, late); err == nil { t.Fatal("old crew receipt released the new permit") }
	unchanged, _ := permits.Get(newPermit.ID)
	if unchanged.Phase != model.PhaseWorking || unchanged.Owner != "night" { t.Fatalf("new permit changed after stale receipt: %+v", unchanged) }
	current, _ := crews.Current("WS-1")
	valid := crew.HandoverReceipt(current, newPermit)
	if _, err := permits.AcceptRelease(crews, valid); err != nil { t.Fatalf("current handover receipt was rejected: %v", err) }
}
