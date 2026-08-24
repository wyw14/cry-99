package api

import (
	"net/http"
	"time"
)

func (a *Application) listEnergize(w http.ResponseWriter, r *http.Request) {
	permitID := r.URL.Query().Get("permit_id")
	writeJSON(w, http.StatusOK, map[string]any{"readiness": a.Readiness.View(permitID), "commands": a.Commands.All()})
}

func (a *Application) createEnergize(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PermitID string   `json:"permit_id"`
		AreaID   string   `json:"area_id"`
		Devices  []string `json:"devices"`
	}
	if err := decodeJSON(r, &request); err != nil || request.PermitID == "" || request.AreaID == "" || len(request.Devices) == 0 {
		writeError(w, http.StatusBadRequest, "permit_id, area_id and devices are required")
		return
	}
	if err := a.Eligibility.Ready(request.AreaID, request.PermitID, time.Now().UTC()); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	operation := a.Sequencer.Plan(request.PermitID, request.Devices)
	result := a.Runner.Run(r.Context(), request.PermitID)
	select {
	case err := <-result:
		if err != nil {
			writeError(w, http.StatusRequestTimeout, err.Error())
			return
		}
		operation, _ = a.Sequencer.Get(request.PermitID)
		for _, command := range operation.Commands {
			a.Evidence.Put(command)
			a.Journal.RecordCommand(command)
		}
		for _, item := range a.Alarms.ForPermit(request.PermitID) {
			a.Journal.RecordAlarm(item)
		}
		writeJSON(w, http.StatusAccepted, operation)
	case <-r.Context().Done():
		a.Runner.Cancel(request.PermitID)
		writeError(w, http.StatusRequestTimeout, r.Context().Err().Error())
	}
}
