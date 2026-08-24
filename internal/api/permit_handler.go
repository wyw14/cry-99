package api

import (
	"net/http"

	"example.com/railvolt/internal/crew"
	"github.com/go-chi/chi/v5"
)

func (a *Application) listPermits(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"permits": a.Permits.List()})
}

func (a *Application) createPermit(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Worksite string `json:"worksite"`
		Owner    string `json:"owner"`
	}
	if err := decodeJSON(r, &request); err != nil || request.Worksite == "" || request.Owner == "" {
		writeError(w, http.StatusBadRequest, "worksite and owner are required")
		return
	}
	session := a.Crew.Handover(request.Worksite, request.Owner, 1)
	p := a.Permits.Create(request.Worksite, request.Owner, a.Topology.Revision())
	receipt := crew.HandoverReceipt(session, p)
	a.Journal.RecordPermit(p)
	writeJSON(w, http.StatusCreated, map[string]any{"permit": p, "handover": receipt})
}

func (a *Application) releasePermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "permitID")
	p, ok := a.Permits.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "permit not found")
		return
	}
	session, ok := a.Crew.Current(p.Worksite)
	if !ok {
		writeError(w, http.StatusConflict, "crew session missing")
		return
	}
	receipt := crew.HandoverReceipt(session, p)
	released, err := a.Permits.AcceptRelease(a.Crew, receipt)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	a.Journal.RecordPermit(released)
	a.Permits.LockEnergize(released.ID, false)
	a.PermitHistory.Record(released, p.Phase, released.Phase, session.Owner, "crew release receipt accepted")
	writeJSON(w, http.StatusOK, released)
}
