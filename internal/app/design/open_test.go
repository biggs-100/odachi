package design_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	appdesign "github.com/biggs-100/odachi/internal/app/design"
)

// The address a clone records survives a round trip, and a file that says nothing is
// not an address: the next open starts its own canvas instead of trusting a ghost.
func TestTheRecordedAddressRoundTripsAndABrokenOneIsNotTrusted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "design-canvas.json")

	canvas := appdesign.Canvas{Addr: "127.0.0.1:9999", URL: "http://127.0.0.1:9999/", Slug: "pricing", PID: 42, StartedAt: "2026-11-01T00:00:00Z"}
	if err := appdesign.WriteCanvas(path, canvas); err != nil {
		t.Fatalf("WriteCanvas: %v", err)
	}
	back, ok := appdesign.ReadCanvas(path)
	if !ok || back.Addr != canvas.Addr || back.Slug != canvas.Slug {
		t.Fatalf("ReadCanvas = %+v, %v", back, ok)
	}

	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := appdesign.ReadCanvas(path); ok {
		t.Fatal("a file that is not an address was trusted")
	}

	if err := appdesign.ForgetCanvas(path); err != nil {
		t.Fatalf("ForgetCanvas: %v", err)
	}
	if _, ok := appdesign.ReadCanvas(path); ok {
		t.Fatal("the address survived forget")
	}
	if err := appdesign.ForgetCanvas(path); err != nil {
		t.Fatalf("forgetting twice is success, got %v", err)
	}
}

// Running asks the server, not the file: a canvas whose terminal died leaves an
// address that answers nothing, and that is exactly what this call is for.
func TestRunningAsksTheServer(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	alive := appdesign.Canvas{Addr: up.Listener.Addr().String()}
	if !appdesign.Running(nil, alive) {
		t.Fatalf("a server that answers health is not running")
	}

	dead := httptest.NewServer(http.NotFoundHandler())
	addr := dead.Listener.Addr().String()
	dead.Close()
	if appdesign.Running(nil, appdesign.Canvas{Addr: addr}) {
		t.Fatalf("a closed port reported as running")
	}
}

// Stop asks and then forgets; a server that cannot be reached is an error and keeps
// the address, so the person can see what it pointed at.
func TestStopForgetsOnlyWhatItReached(t *testing.T) {
	path := filepath.Join(t.TempDir(), "design-canvas.json")

	stopped := false
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopped = true
		w.WriteHeader(http.StatusOK)
	}))
	canvas := appdesign.Canvas{Addr: up.Listener.Addr().String()}
	if err := appdesign.WriteCanvas(path, canvas); err != nil {
		t.Fatal(err)
	}
	if err := appdesign.Stop(nil, path, canvas); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	up.Close()
	if !stopped {
		t.Fatal("Stop did not ask the canvas")
	}
	if _, ok := appdesign.ReadCanvas(path); ok {
		t.Fatal("the address survived a stop that was heard")
	}
}
