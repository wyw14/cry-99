package api

import (
	"context"
	"net/http"
	"time"
)

type Server struct{ http *http.Server }

func NewServer(address string, app *Application) *Server {
	return &Server{http: &http.Server{Addr: address, Handler: app.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}}
}
func (s *Server) ListenAndServe() error              { return s.http.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }
