package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *Application) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Get("/healthz", a.health)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/permits", http.StatusTemporaryRedirect)
	})
	r.Get("/permits", a.page("permits.html"))
	r.Get("/isolation", a.page("isolation.html"))
	r.Get("/energize", a.page("energize.html"))
	r.Get("/events", a.page("events.html"))
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(a.WebDir))))
	r.Route("/api", func(r chi.Router) {
		r.Get("/permits", a.listPermits)
		r.Post("/permits", a.createPermit)
		r.Post("/permits/{permitID}/release", a.releasePermit)
		r.Get("/isolation", a.listIsolation)
		r.Post("/isolation", a.createIsolation)
		r.Post("/isolation/{planID}/step", a.stepIsolation)
		r.Post("/isolation/{planID}/confirm", a.confirmIsolation)
		r.Post("/isolation/{planID}/cancel", a.cancelIsolation)
		r.Get("/energize", a.listEnergize)
		r.Post("/energize", a.createEnergize)
		r.Get("/events", a.listEvents)
		r.Get("/alarms", a.listAlarms)
		r.Get("/topology", a.getTopology)
		r.Post("/topology", a.publishTopology)
		r.Post("/topology/{areaID}/split", a.splitTopology)
		r.Get("/ground", a.listGrounds)
		r.Post("/ground", a.applyGround)
		r.Post("/ground/{groundID}/verify", a.verifyGround)
		r.Post("/ground/{groundID}/release", a.releaseGround)
		r.Post("/recovery/checkpoint", a.createCheckpoint)
		r.Get("/recovery", a.recoveryStatus)
		r.Get("/status", a.systemStatus)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { writeError(w, http.StatusNotFound, "route not found") })
	return r
}
