package testcert_test

import (
	"testing"
	"time"

	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/testcert"
)

// TestSplitRejectsStaleSupersetCertificate reproduces the maintenance-split
// defect: a supply arm is split into two energization sections and only one is
// re-tested. The stale certificate covering the original (larger) boundary set
// must not satisfy either smaller section -- each section needs a certificate
// whose scope matches its current boundaries.
func TestSplitRejectsStaleSupersetCertificate(t *testing.T) {
	registry := testcert.NewRegistry()
	now := time.Now().UTC()
	ttl := 8 * time.Hour

	// Original supply arm: one section spanning four boundaries, tested whole.
	original := model.Area{ID: "ARM", Boundaries: []string{"b1", "b2", "b3", "b4"}}
	registry.Issue(original.ID, "rev-1", original.Boundaries, true, ttl)

	// Maintenance splits the arm. The operator keeps the ARM id for the first
	// section but narrows its boundaries, and only re-tests the second section.
	armSection := model.Area{ID: "ARM", Boundaries: []string{"b1", "b2"}}
	otherSection := model.Area{ID: "ARM2", Boundaries: []string{"b3", "b4"}}
	registry.Issue(otherSection.ID, "rev-2", otherSection.Boundaries, true, ttl)

	// The un-retested ARM section must NOT be covered by the stale whole-arm
	// certificate, even though that certificate's scope is a superset.
	if _, ok := registry.ValidForArea(armSection, "", now); ok {
		t.Fatalf("stale superset certificate must not cover split section ARM")
	}

	// The re-tested section is covered by its own current-scope certificate.
	if _, ok := registry.ValidForArea(otherSection, "", now); ok == false {
		t.Fatalf("current-scope certificate must cover re-tested section ARM2")
	}

	// Once ARM is re-tested against its narrowed scope, it becomes eligible too.
	registry.Issue(armSection.ID, "rev-2", armSection.Boundaries, true, ttl)
	if _, ok := registry.ValidForArea(armSection, "", now); ok == false {
		t.Fatalf("current-scope certificate must cover re-tested section ARM")
	}
}
