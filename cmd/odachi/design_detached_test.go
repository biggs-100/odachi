package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestDesignOpenReturnsWithADetachedCanvas proves the child outlives the command:
// `open` returns while the canvas keeps answering, a second `open` moves the same
// server instead of starting another, and `stop` closes it. It builds the real
// binary because the detached half is the binary spawning itself.
func TestDesignOpenReturnsWithADetachedCanvas(t *testing.T) {
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	binary := filepath.Join(t.TempDir(), "odachi")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/odachi")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	dir := t.TempDir()
	design := filepath.Join(dir, "design", "pricing")
	if err := os.MkdirAll(design, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"design.json": `{"title":"Pricing","brief":"b","artboards":[{"file":"01-a.html","title":"A","tradeoff":"dense","width":800,"height":600}]}`,
		"01-a.html":   "<!doctype html><h1>Ship faster</h1>",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(design, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git", "odachi"), 0o755); err != nil {
		t.Fatal(err)
	}

	stop := func() {
		command := exec.Command(binary, "design", "stop")
		command.Dir = dir
		_ = command.Run()
	}
	t.Cleanup(stop)

	started := time.Now()
	open := exec.Command(binary, "design", "open", "--no-browser")
	open.Dir = dir
	if out, err := open.CombinedOutput(); err != nil {
		t.Fatalf("open: %v\n%s", err, out)
	}
	if elapsed := time.Since(started); elapsed > 10*time.Second {
		t.Fatalf("open blocked for %s; the canvas is not detached", elapsed)
	}

	statePath := filepath.Join(dir, ".git", "odachi", "design-canvas.json")
	canvas := readCanvasState(t, statePath)
	client := &http.Client{Timeout: time.Second}
	health := "http://" + canvas.Addr + "/" + canvas.Token + "/api/health"
	response, err := client.Get(health)
	if err != nil {
		t.Fatalf("the canvas does not answer after open: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health = %d, want 200", response.StatusCode)
	}

	// A second open moves the page that already exists; it does not start a server.
	again := exec.Command(binary, "design", "open", "--no-browser", "pricing")
	again.Dir = dir
	if out, err := again.CombinedOutput(); err != nil {
		t.Fatalf("second open: %v\n%s", err, out)
	}
	if moved := readCanvasState(t, statePath); moved.Addr != canvas.Addr {
		t.Fatalf("a second open started another canvas: %s then %s", canvas.Addr, moved.Addr)
	}

	command := exec.Command(binary, "design", "stop")
	command.Dir = dir
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("stop: %v\n%s", err, out)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := client.Get(health); err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("the canvas still answers after stop; log:\n%s", canvasLog(t, dir))
}

func readCanvasState(t *testing.T, path string) struct {
	Addr  string `json:"addr"`
	URL   string `json:"url"`
	Token string `json:"token"`
} {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the canvas state: %v", err)
	}
	var state struct {
		Addr  string `json:"addr"`
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("decode the canvas state: %v", err)
	}
	return state
}

func canvasLog(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".git", "odachi", "design-canvas.log"))
	if err != nil {
		return "(no log)"
	}
	return string(raw)
}
