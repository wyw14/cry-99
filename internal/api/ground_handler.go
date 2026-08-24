package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (a *Application) listGrounds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"records": a.Grounds.Records(), "checklists": a.GroundChecklist.Grounds()})
}

func (a *Application) applyGround(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AreaID string `json:"area_id"`
	}
	if err := decodeJSON(r, &request); err != nil || request.AreaID == "" {
		writeError(w, http.StatusBadRequest, "area_id is required")
		return
	}
	id := a.Grounds.Apply(request.AreaID)
	checks, _ := a.GroundChecklist.Begin(id, request.AreaID)
	a.GroundLock.Set(request.AreaID, true)
	a.RefreshGroundView()
	writeJSON(w, http.StatusCreated, map[string]any{"ground_id": id, "checks": checks})
}

func (a *Application) verifyGround(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groundID")
	if err := a.Grounds.Verify(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	for _, check := range a.GroundChecklist.Get(id) {
		if check.Name == "device_identity" || check.Name == "physical_position" || check.Name == "zero_current" {
			a.GroundChecklist.Decide(id, check.ID, true, "field-team", "verified")
		}
	}
	a.RefreshGroundView()
	writeJSON(w, http.StatusOK, map[string]any{"ground_id": id, "verified": true})
}

func (a *Application) releaseGround(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groundID")
	evidence, ok := a.Grounds.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "ground not found")
		return
	}
	if !a.TelemetryWindow.StableZero(evidence.AreaID, time.Second, time.Now().UTC().Add(time.Second)) {
		writeError(w, http.StatusConflict, "zero current window is not stable")
		return
	}
	released, err := a.Grounds.Release(id)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if !a.Grounds.Released(id) {
		writeError(w, http.StatusConflict, "ground release did not become durable")
		return
	}
	for _, check := range a.GroundChecklist.Get(id) {
		if check.Name == "evidence_persisted" {
			a.GroundChecklist.Decide(id, check.ID, true, "system", "durable telemetry recorded")
		}
	}
	a.GroundLock.Set(evidence.AreaID, false)
	a.RefreshGroundView()
	writeJSON(w, http.StatusOK, released)
}
