package design

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Canvas is the address of a running canvas: the file that says one is up, so a
// second `open` refreshes the page a person already has instead of starting a
// second server.
type Canvas struct {
	Addr      string `json:"addr"`
	URL       string `json:"url"`
	Slug      string `json:"slug"`
	PID       int    `json:"pid"`
	StartedAt string `json:"startedAt"`
}

// CanvasPath is where the clone keeps the address: the store, next to the verdicts,
// because it is state about this checkout and never travels with the tree.
func CanvasPath(dir string) string {
	return filepath.Join(dir, ".git", "odachi", "design-canvas.json")
}

// ReadCanvas returns the address a clone recorded, and whether there is one. A file
// that cannot be read is not a canvas: the next open starts its own.
func ReadCanvas(path string) (Canvas, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Canvas{}, false
	}
	var canvas Canvas
	if err := json.Unmarshal(raw, &canvas); err != nil || strings.TrimSpace(canvas.Addr) == "" {
		return Canvas{}, false
	}
	return canvas, true
}

// WriteCanvas records the address, creating the store when the clone has none.
func WriteCanvas(path string, canvas Canvas) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(canvas, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

// ForgetCanvas removes the address. It runs when the server stops, and a missing
// file is success: the state is the address, and no address is not an error.
func ForgetCanvas(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Running asks the recorded canvas whether it is still there. A server that died
// without cleaning up leaves a file, and trusting a file instead of asking it is how
// a person gets a URL that answers nothing.
func Running(client *http.Client, canvas Canvas) bool {
	if client == nil {
		client = &http.Client{Timeout: 800 * time.Millisecond}
	}
	response, err := client.Get("http://" + canvas.Addr + "/api/health")
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

// Stop asks the canvas to close, and forgets the address only when the server heard
// it: a stop that removes the file of a live server loses the only way to reach it.
func Stop(client *http.Client, path string, canvas Canvas) error {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	response, err := client.Post("http://"+canvas.Addr+"/api/stop", "application/json", nil)
	if err != nil {
		return fmt.Errorf("could not reach the canvas at %s: %w", canvas.Addr, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("the canvas at %s answered %s", canvas.Addr, response.Status)
	}
	return ForgetCanvas(path)
}
