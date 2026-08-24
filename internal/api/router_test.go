package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/railvolt/internal/api"
)

func TestOperationalRoutes(t *testing.T) {
	web := t.TempDir()
	for _, name := range []string{"permits.html", "isolation.html", "energize.html", "events.html"} {
		if err := os.WriteFile(filepath.Join(web, name), []byte("<html>"+name+"</html>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	app, err := api.NewApplication(t.TempDir(), web)
	if err != nil {
		t.Fatal(err)
	}
	handler := app.Router()
	for _, path := range []string{"/permits", "/isolation", "/energize", "/events", "/healthz", "/api/permits", "/api/isolation", "/api/energize", "/api/status"} {
		t.Run(strings.TrimPrefix(path, "/"), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("GET %s returned %d: %s", path, response.Code, response.Body.String())
			}
		})
	}
}

func TestPermitCreationUsesTopologyRevision(t *testing.T) {
	app, err := api.NewApplication(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	body := strings.NewReader(`{"worksite":"WS-22","owner":"night"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/permits", body)
	response := httptest.NewRecorder()
	app.Router().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), app.Topology.Revision()) {
		t.Fatalf("permit response does not bind the active topology revision")
	}
}
