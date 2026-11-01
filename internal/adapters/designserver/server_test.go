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
	"time"
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

func canvas(t *testing.T, root string) string {
	t.Helper()
	server := New(Options{Dir: root, Canvas: []byte("<!doctype html><title>canvas</title>")})
	addr, _, err := server.Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return "http://" + addr
}

func TestTheCanvasServesTheDesignsAndTheirArtboards(t *testing.T) {
	base := canvas(t, project(t))

	var designs struct {
		Designs []Design `json:"designs"`
	}
	get(t, base+"/api/designs", &designs)
	if len(designs.Designs) != 1 || designs.Designs[0].Slug != "pricing" {
		t.Fatalf("designs = %+v", designs.Designs)
	}
	if designs.Designs[0].URL["01-a.html"] != "/artboard/pricing/01-a.html" {
		t.Errorf("the artboard has no address: %+v", designs.Designs[0].URL)
	}

	body := getText(t, base+"/artboard/pricing/01-a.html")
	if !strings.Contains(body, "Ship faster") {
		t.Fatalf("the artboard is not served: %q", body)
	}
}

// The canvas is not a file browser: only a file the manifest lists is served, and a
// path that tries to leave the folder is nothing at all.
func TestTheCanvasServesOnlyWhatTheManifestLists(t *testing.T) {
	base := canvas(t, project(t))

	for _, path := range []string{
		"/artboard/pricing/design.json",
		"/artboard/pricing/..%2Fdesign.json",
		"/artboard/pricing/nope.html",
	} {
		response, err := httpGet(base + path)
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
	base := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","excerpt":"Ship faster","comment":"Shorter headline."}`)
	response, err := http.Post(base+"/api/design/pricing/comments", "application/json", body)
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
	base := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"Shorter."}`)
	created, err := http.Post(base+"/api/design/pricing/comments", "application/json", body)
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

	response, err := http.Post(base+"/api/design/pricing/comments/"+comment.ID, "application/json", nil)
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
	base := canvas(t, root)

	body := strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"Shorter."}`)
	response, err := http.Post(base+"/api/design/pricing/comments", "application/json", body)
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

// The canvas answers its own page and nothing else: a page from the web must not
// append a comment or stop the canvas just because the port answers on loopback.
func TestACrossOriginRequestIsRefused(t *testing.T) {
	base := canvas(t, project(t))

	request, err := http.NewRequest(http.MethodPost, base+"/api/design/pricing/comments", strings.NewReader(`{"file":"01-a.html","comment":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://evil.example")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin POST = %d, want 403", response.StatusCode)
	}

	stop, err := http.NewRequest(http.MethodPost, base+"/api/stop", nil)
	if err != nil {
		t.Fatal(err)
	}
	stop.Header.Set("Origin", "http://evil.example")
	stopped, err := http.DefaultClient.Do(stop)
	if err != nil {
		t.Fatal(err)
	}
	stopped.Body.Close()
	if stopped.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin stop = %d, want 403", stopped.StatusCode)
	}
}

// A body that is not JSON is refused before it is read: a simple text/plain form
// post needs no preflight, and that is the door the origin check cannot watch.
func TestABodyThatIsNotJSONIsRefused(t *testing.T) {
	base := canvas(t, project(t))

	response, err := http.Post(base+"/api/design/pricing/comments", "text/plain", strings.NewReader(`{"file":"01-a.html","comment":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("text/plain POST = %d, want 415", response.StatusCode)
	}
}

// The watcher announces every file that moved in one poll: an agent applying two
// comments writes two artboards, and the second must not swallow the first.
func TestTheWatcherAnnouncesEveryFileThatMoved(t *testing.T) {
	root := project(t)
	dir := filepath.Join(root, "design", "pricing")
	if err := os.WriteFile(filepath.Join(dir, "02-b.html"), []byte("<h1>b</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := New(Options{Dir: root, Canvas: []byte("canvas")})
	server.changed() // seed: the first look records what is already there

	later := time.Now().Add(2 * time.Second)
	for _, name := range []string{"01-a.html", "02-b.html"} {
		path := filepath.Join(dir, name)
		if err := os.Chtimes(path, later, later); err != nil {
			t.Fatal(err)
		}
	}

	moved := server.changed()
	if len(moved) != 2 {
		t.Fatalf("moved = %v, want the two artboards", moved)
	}
}

// The origin check is the door, so both halves are pinned against a server that
// really listened: the canvas accepts its own page and refuses a foreign one, and a
// canvas that does not know its address refuses instead of trusting the header it was
// given. The address is set by Listen, never by the test, so the run is race-free.
func TestOwnPageAcceptsItsOwnOriginAndRefusesTheRest(t *testing.T) {
	server := New(Options{Dir: project(t), Canvas: []byte("canvas")})
	addr, _, err := server.Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer server.Close()

	// The positive case takes a comment, not the stop route: a canvas this test
	// killed could not be asked the next question.
	post := func(origin string) int {
		request, err := http.NewRequest(http.MethodPost, "http://"+addr+"/api/design/pricing/comments",
			strings.NewReader(`{"file":"01-a.html","comment":"from the canvas"}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		if origin != "" {
			request.Header.Set("Origin", origin)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		return response.StatusCode
	}

	if got := post("http://" + addr); got != http.StatusCreated {
		t.Errorf("the canvas's own origin = %d, want 201", got)
	}
	if got := post("http://evil.example"); got != http.StatusForbidden {
		t.Errorf("a foreign origin on the comments route = %d, want 403", got)
	}

	// The stop route is a door too, and it is asked last: it is the one that ends
	// the canvas.
	stop, err := http.NewRequest(http.MethodPost, "http://"+addr+"/api/stop", nil)
	if err != nil {
		t.Fatal(err)
	}
	stop.Header.Set("Origin", "http://evil.example")
	stopped, err := http.DefaultClient.Do(stop)
	if err != nil {
		t.Fatal(err)
	}
	stopped.Body.Close()
	if stopped.StatusCode != http.StatusForbidden {
		t.Errorf("a foreign origin on the stop route = %d, want 403", stopped.StatusCode)
	}

	// A canvas that never listened does not know its address: it refuses, with and
	// without an Origin header, instead of trusting what the caller says.
	blind := New(Options{Dir: project(t), Canvas: []byte("canvas")})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/stop", nil)
	blind.stop(response, request)
	if response.Code != http.StatusForbidden {
		t.Errorf("a canvas that does not know its address = %d, want 403", response.Code)
	}

	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/stop", nil)
	request.Header.Set("Origin", "http://whatever.example")
	blind.stop(response, request)
	if response.Code != http.StatusForbidden {
		t.Errorf("a canvas that does not know its address with an origin = %d, want 403", response.Code)
	}
}
