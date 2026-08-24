package api

import (
	"net/http"
	"time"
)

func (a *Application) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC(), "revision": a.Topology.Revision()})
}
