package api

import "net/http"

func (a *Application) listEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"revision": a.Journal.Revision(), "events": a.Journal.Events()})
}
func (a *Application) listAlarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"alarms": a.Alarms.List()})
}
func (a *Application) getTopology(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.Topology.Current())
}
