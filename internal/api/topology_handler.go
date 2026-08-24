package api

import (
	"net/http"

	"example.com/railvolt/internal/model"
	"example.com/railvolt/internal/topology"
	"github.com/go-chi/chi/v5"
)

func (a *Application) publishTopology(w http.ResponseWriter, r *http.Request) {
	var request model.Topology
	if err := decodeJSON(r, &request); err != nil || len(request.Areas) == 0 {
		writeError(w, http.StatusBadRequest, "topology with areas is required")
		return
	}
	before := a.Topology.Current()
	revision := a.Topology.Publish(request)
	after := a.Topology.Current()
	a.Changes.Record(topology.ChangeReturnGroup, before, after, map[string]string{"source": "http"})
	a.CertificateFlow.ArchiveForRevision(revision, "topology revision changed")
	a.RefreshGroundView()
	writeJSON(w, http.StatusOK, after)
}

func (a *Application) splitTopology(w http.ResponseWriter, r *http.Request) {
	var request struct {
		First  model.Area `json:"first"`
		Second model.Area `json:"second"`
	}
	if err := decodeJSON(r, &request); err != nil || request.First.ID == "" || request.Second.ID == "" {
		writeError(w, http.StatusBadRequest, "two replacement areas are required")
		return
	}
	before := a.Topology.Current()
	revision, err := a.Topology.SplitArea(chi.URLParam(r, "areaID"), request.First, request.Second)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	after := a.Topology.Current()
	a.Changes.Record(topology.ChangeSplit, before, after, map[string]string{"area": chi.URLParam(r, "areaID")})
	a.CertificateFlow.ArchiveForRevision(revision, "area split requires current-scope evidence")
	a.RefreshGroundView()
	writeJSON(w, http.StatusOK, after)
}
