package api

import (
	"net/http"

	"example.com/railvolt/internal/isolation"
	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/switchgear"
	"github.com/go-chi/chi/v5"
)

func (a *Application) listIsolation(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"plans": a.Plans.List()})
}

func (a *Application) createIsolation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PermitID string   `json:"permit_id"`
		Devices  []string `json:"devices"`
	}
	if err := decodeJSON(r, &request); err != nil || request.PermitID == "" || len(request.Devices) == 0 {
		writeError(w, http.StatusBadRequest, "permit_id and devices are required")
		return
	}
	permitValue, ok := a.Permits.Get(request.PermitID)
	if !ok {
		writeError(w, http.StatusNotFound, "permit not found")
		return
	}
	previousPhase := permitValue.Phase
	if previousPhase == model.PhaseDraft {
		if advanced, err := a.Permits.Advance(request.PermitID, model.PhaseIsolating); err == nil {
			a.PermitHistory.Record(advanced, previousPhase, advanced.Phase, "duty-controller", "isolation plan started")
		}
	}
	for _, deviceID := range request.Devices {
		if _, exists := a.Fleet.Get(deviceID); !exists {
			a.Fleet.Register(switchgear.Device{ID: deviceID, Kind: "disconnect", AreaID: permitValue.Worksite, Position: "closed", RemoteEnabled: true})
		}
	}
	plan := a.Isolation.Begin(request.PermitID, request.Devices)
	writeJSON(w, http.StatusCreated, plan)
}

func (a *Application) stepIsolation(w http.ResponseWriter, r *http.Request) {
	plan, err := a.Isolation.ExecuteNext(r.Context(), chi.URLParam(r, "planID"))
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (a *Application) confirmIsolation(w http.ResponseWriter, r *http.Request) {
	planID := chi.URLParam(r, "planID")
	before, _ := a.Isolation.Plan(planID)
	plan, err := a.Isolation.Confirm(planID)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if before.Current < len(before.Steps) {
		step := before.Steps[before.Current]
		if device, ok := a.Fleet.Get(step.DeviceID); ok {
			a.Fleet.UpdatePosition(step.DeviceID, "open", device.Epoch+1)
		}
	}
	if isolation.Complete(plan) {
		a.Permits.UnlockForWork(plan.PermitID)
	}
	writeJSON(w, http.StatusOK, plan)
}

func (a *Application) cancelIsolation(w http.ResponseWriter, r *http.Request) {
	recovery, err := a.Isolation.Cancel(chi.URLParam(r, "planID"), a.Permits)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	for _, step := range recovery.InFlight {
		a.Fleet.SetLocks(step.DeviceID, true, false)
	}
	writeJSON(w, http.StatusOK, recovery)
}
