package api

import (
	"net/http"
	"path/filepath"
)

func (a *Application) page(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(a.WebDir, name)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, path)
	}
}
