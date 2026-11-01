package designserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// project builds a design folder a canvas can serve.
func project(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	dir := filepath.Join(root, "design", "pricing")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"design.json": `{"title":"Pricing page","brief":"b","artboards":[{"file":"01-a.html","title":"A","tradeoff":"dense","width":1440,"height":900}]}`,
		"01-a.html":   "<!doctype html><h1>Ship faster</h1>",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func canvas(t *testing.T, root string) *httptest.Server {
	t.Helper()
	server := New(Options{Dir: root, Canvas: []byte("<!doctype html><title>canvas</title>")})
	srv := httptest.NewServer(server.routes())
	t.Cleanup(srv.Close)
	return srv
}

func TestTheCanvasServesTheDesignsAndTheirArtboards(t *testing.T) {
	srv := canvas(t, project(t))

	var designs struct {
		Designs []Design `json:"designs"`
	}
	get(t, srv.URL+"/api/designs", &designs)
	if len(designs.Designs) != 1 || designs.Designs[0].Slug != "pricing" {
		t.Fatalf("designs = %+v", designs.Designs)
	}
	if designs.Designs[0].URL["01-a.html"] != "/artboard/pricing/01-a.html" {
		t.Errorf("the artboard has no address: %+v", designs.Designs[0].URL)
	}

	body := getText(t, srv.URL+"/artboard/pricing/01-a.html")
	if !strings.Contains(body, "Ship faster") {
		t.Fatalf("the artboard is not served: %q", body)
	}
}

// The canvas is not a file browser: only a file the manifest lists is served, and a
// path that tries to leave the folder is nothing at all.
func TestTheCanvasServesOnlyWhatTheManifestLists(t *testing.T) {
	srv := canvas(t, project(t))

	for _, path := range []string{
		"/artboard/pricing/design.json",
		"/artboard/pricing/..%2Fdesign.json",
		"/artboard/pricing/nope.html",
	} {
		response, err := httpGet(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, response.StatusCode)
		}
	}
}

// A comment is written to the file every agent reads, with an id the skill documents
// and the state the page and the command agree on.
func TestACommentIsWrittenWhereEveryAgentReadsIt(t *testing.T) {
	root := project(t)
	srv := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","excerpt":"Ship faster","comment":"Shorter headline."}`)
	response, err := http.Post(srv.URL+"/api/design/pricing/comments", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST comment = %d", response.StatusCode)
	}

	raw, err := os.ReadFile(filepath.Join(root, "design", "pricing", "comments.json"))
	if err != nil {
		t.Fatalf("the canvas did not write the comments file: %v", err)
	}
	var file struct {
		Version  int `json:"version"`
		Comments []struct {
			ID      string `json:"id"`
			Comment string `json:"comment"`
			Status  string `json:"status"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("comments.json is not JSON: %v\n%s", err, raw)
	}
	if file.Version != 1 || len(file.Comments) != 1 || file.Comments[0].Status != "open" || len(file.Comments[0].ID) != 8 {
		t.Fatalf("comments.json = %+v", file)
	}
}

// Resolving is a POST on the comment and it keeps the file valid.
func TestResolvingACommentKeepsTheFileValid(t *testing.T) {
	root := project(t)
	srv := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"Shorter."}`)
	created, err := http.Post(srv.URL+"/api/design/pricing/comments", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	var comment struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&comment); err != nil {
		t.Fatal(err)
	}
	created.Body.Close()

	response, err := http.Post(srv.URL+"/api/design/pricing/comments/"+comment.ID, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("resolve = %d", response.StatusCode)
	}

	raw, err := os.ReadFile(filepath.Join(root, "design", "pricing", "comments.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"status": "resolved"`) {
		t.Fatalf("the comment was not resolved:\n%s", raw)
	}
}

// A canvas that cannot parse the comments file refuses to write it: destroying a
// person's notes is worse than refusing a click.
func TestACanvasNeverOverwritesAFileItCannotParse(t *testing.T) {
	root := project(t)
	comments := filepath.Join(root, "design", "pricing", "comments.json")
	if err := os.WriteFile(comments, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"Shorter."}`)
	response, err := http.Post(srv.URL+"/api/design/pricing/comments", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusConflict {
		t.Fatalf("POST over a broken file = %d, want 409", response.StatusCode)
	}
	raw, err := os.ReadFile(comments)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "{" {
		t.Fatalf("the canvas rewrote a file it could not parse: %q", raw)
	}
}

func get(t *testing.T, url string, into any) {
	t.Helper()
	response, err := httpGet(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(into); err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
}

func getText(t *testing.T, url string) string {
	t.Helper()
	response, err := httpGet(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return string(body)
}

func httpGet(url string) (*http.Response, error) {
	return http.Get(url)
}
