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
	_, url, err := server.Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return strings.TrimSuffix(url, "/")
}

// A second `open` moves the tab a person already has: the server records the design
// and tells every open page to re-render, instead of serving a second canvas.
func TestShowMovesTheCanvasToAKnownDesign(t *testing.T) {
	base := canvas(t, project(t))

	response, err := http.Post(base+"/api/show", "application/json", strings.NewReader(`{"slug":"pricing"}`))
	if err != nil {
		t.Fatalf("POST show: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("POST show = %d, want 200", response.StatusCode)
	}

	response, err = http.Post(base+"/api/show", "application/json", strings.NewReader(`{"slug":"nope"}`))
	if err != nil {
		t.Fatalf("POST show unknown: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("POST show unknown = %d, want 404", response.StatusCode)
	}
}

// The design a page shows is a control like the rest: a foreign origin cannot move
// it, and a body that is not JSON cannot either.
func TestShowRefusesAForeignOriginAndANonJSONBody(t *testing.T) {
	base := canvas(t, project(t))

	request, err := http.NewRequest(http.MethodPost, base+"/api/show", strings.NewReader(`{"slug":"pricing"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://evil.example")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("POST show: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("foreign origin = %d, want 403", response.StatusCode)
	}

	response, err = http.Post(base+"/api/show", "text/plain", strings.NewReader(`{"slug":"pricing"}`))
	if err != nil {
		t.Fatalf("POST show text: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("text body = %d, want 415", response.StatusCode)
	}
}

// The canvas has one secret in its address: a route without the token does not
// exist, and a request that names another host is refused before it reaches the mux.
func TestTheCanvasRefusesARouteWithoutItsToken(t *testing.T) {
	base := canvas(t, project(t))
	host, _, _ := strings.Cut(strings.TrimPrefix(base, "http://"), "/")

	response, err := http.Get("http://" + host + "/api/designs")
	if err != nil {
		t.Fatalf("GET without token: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("route without token = %d, want 404", response.StatusCode)
	}

	request, err := http.NewRequest(http.MethodGet, base+"/api/designs", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "evil.example"
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET foreign host: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("foreign host = %d, want 403", response.StatusCode)
	}
}

// The artboard is served with the picker that names a clicked element: the canvas
// cannot read a sandboxed page, so the page has to say what was clicked, where its
// comments are, and when Shift+Esc asks to leave focus.
func TestTheArtboardCarriesTheSandboxedPicker(t *testing.T) {
	base := canvas(t, project(t))
	body := getText(t, base+"/artboard/pricing/01-a.html")
	for _, want := range []string{"odachi:pick", "odachi:pins", "odachi:blur"} {
		if !strings.Contains(body, want) {
			t.Errorf("the artboard does not carry %q: %q", want, body)
		}
	}
}

// A comment is a note to an agent, not a document: the server bounds it instead of
// writing whatever a page sends.
func TestACommentIsBounded(t *testing.T) {
	base := canvas(t, project(t))
	long := strings.Repeat("x", 4001)
	response, err := http.Post(base+"/api/design/pricing/comments", "application/json",
		strings.NewReader(`{"file":"01-a.html","comment":"`+long+`"}`))
	if err != nil {
		t.Fatalf("POST long comment: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("a 4001-byte comment = %d, want 400", response.StatusCode)
	}
}

// A symlink inside the design folder can point outside it: the served file must
// still resolve inside the design, or the canvas is a file browser with a token.
func TestAnArtboardSymlinkCannotEscapeTheDesign(t *testing.T) {
	root := project(t)
	secret := filepath.Join(root, "secret.html")
	if err := os.WriteFile(secret, []byte("<h1>secret</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "design", "pricing", "02-link.html")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	manifest := filepath.Join(root, "design", "pricing", "design.json")
	raw := `{"title":"Pricing page","brief":"b","artboards":[{"file":"01-a.html","title":"A","tradeoff":"dense","width":1440,"height":900},{"file":"02-link.html","title":"Link","tradeoff":"none","width":1440,"height":900}]}`
	if err := os.WriteFile(manifest, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	base := canvas(t, root)
	response, err := http.Get(base + "/artboard/pricing/02-link.html")
	if err != nil {
		t.Fatalf("GET the symlink: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("a symlink that leaves the design = %d, want 404", response.StatusCode)
	}
}

// The system board is the first thing the canvas draws when the project has one:
// the page reads its URL from the same payload as the designs.
func TestTheCanvasServesTheSystemBoard(t *testing.T) {
	root := project(t)
	if err := os.WriteFile(filepath.Join(root, "design", "system.html"), []byte("<h1>tokens</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := canvas(t, root)

	var payload struct {
		System string `json:"system"`
	}
	get(t, base+"/api/designs", &payload)
	if !strings.HasSuffix(payload.System, "/system") {
		t.Fatalf("system = %q, want the board's address", payload.System)
	}
	body := getText(t, base+"/system")
	if !strings.Contains(body, "tokens") {
		t.Errorf("the system board is not served: %q", body)
	}
}

// The canvas reads the handoff state from the folder: the chosen artboard and its
// state are what it draws as a badge.
func TestTheCanvasReportsTheHandoffState(t *testing.T) {
	root := project(t)
	state := `{"version":1,"chosen":"01-a.html","handoffs":{"01-a.html":{"status":"implemented","updatedAt":"t","previewUrl":"http://localhost:3000"}}}`
	if err := os.WriteFile(filepath.Join(root, "design", "pricing", "handoffs.json"), []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}

	base := canvas(t, root)
	var payload struct {
		Designs []Design `json:"designs"`
	}
	get(t, base+"/api/designs", &payload)
	if len(payload.Designs) != 1 {
		t.Fatalf("designs = %+v", payload.Designs)
	}
	handoffs := payload.Designs[0].Handoffs
	if handoffs.Chosen != "01-a.html" || handoffs.Handoffs["01-a.html"].Status != "implemented" {
		t.Errorf("handoffs = %+v", handoffs)
	}
	if handoffs.Handoffs["01-a.html"].PreviewURL != "http://localhost:3000" {
		t.Errorf("the preview did not survive: %+v", handoffs.Handoffs["01-a.html"])
	}
}

func TestTheWatcherAnnouncesAHandoffChange(t *testing.T) {
	root := project(t)
	server := New(Options{Dir: root, Canvas: []byte("canvas")})
	if _, _, err := server.Listen(0); err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.changed()
	if err := os.WriteFile(filepath.Join(root, "design", "pricing", "handoffs.json"), []byte(`{"version":1,"chosen":"01-a.html","handoffs":{"01-a.html":{"status":"sent"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := server.changed()
	if !strings.Contains(strings.Join(moved, ","), "handoffs") {
		t.Errorf("moved = %v, want the handoff change", moved)
	}
}

// A comment sent from the canvas becomes one message file for the session plugin: a
// file per message means the canvas and the plugin never write the same file.
func TestASentCommentIsQueuedForTheSession(t *testing.T) {
	t.Setenv("OPENCODE_SESSION_ID", "s1")
	root := project(t)
	base := canvas(t, root)

	response, err := http.Post(base+"/api/design/pricing/comments", "application/json",
		strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"shorter","send":true}`))
	if err != nil {
		t.Fatalf("POST comment: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST comment = %d, want 201", response.StatusCode)
	}

	entries, err := os.ReadDir(filepath.Join(root, ".git", "odachi", "design-queue"))
	if err != nil {
		t.Fatalf("the queue was not written: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("queue files = %d, want 1", len(entries))
	}
	raw, err := os.ReadFile(filepath.Join(root, ".git", "odachi", "design-queue", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var message struct {
		SessionID string `json:"sessionID"`
		Text      string `json:"text"`
	}
	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("decode the message: %v\n%s", err, raw)
	}
	if message.SessionID != "s1" || !strings.Contains(message.Text, "shorter") {
		t.Errorf("message = %+v", message)
	}
}

// Collect saves without sending; Send all queues every unsent open comment in one
// pass and marks them sent.
func TestCollectedCommentsAreSentInABatch(t *testing.T) {
	t.Setenv("OPENCODE_SESSION_ID", "s1")
	root := project(t)
	base := canvas(t, root)

	for _, text := range []string{"first", "second"} {
		response, err := http.Post(base+"/api/design/pricing/comments", "application/json",
			strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"`+text+`"}`))
		if err != nil {
			t.Fatalf("POST comment: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("POST comment = %d, want 201", response.StatusCode)
		}
	}

	response, err := http.Post(base+"/api/design/pricing/send", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST send: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("POST send = %d, want 200", response.StatusCode)
	}
	var result struct {
		Queued  int  `json:"queued"`
		Session bool `json:"session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Queued != 2 || !result.Session {
		t.Errorf("result = %+v, want 2 queued with a session", result)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".git", "odachi", "design-queue"))
	if err != nil || len(entries) != 2 {
		t.Errorf("queue files = %d (err %v), want 2", len(entries), err)
	}
}

// Without a session the note is still saved: the pull path is the fallback, not a
// lost comment.
func TestACanvasWithoutASessionQueuesNothing(t *testing.T) {
	t.Setenv("OPENCODE_SESSION_ID", "")
	root := project(t)
	base := canvas(t, root)

	response, err := http.Post(base+"/api/design/pricing/comments", "application/json",
		strings.NewReader(`{"file":"01-a.html","selector":"h1","comment":"shorter","send":true}`))
	if err != nil {
		t.Fatalf("POST comment: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST comment = %d, want 201", response.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "odachi", "design-queue")); !os.IsNotExist(err) {
		t.Errorf("a canvas with no session wrote a queue: %v", err)
	}
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
	if !strings.HasSuffix(designs.Designs[0].URL["01-a.html"], "/artboard/pricing/01-a.html") {
		t.Errorf("the artboard has no address under the token: %+v", designs.Designs[0].URL)
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
		request, err := http.NewRequest(http.MethodPost, "http://"+addr+"/"+server.Token()+"/api/design/pricing/comments",
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
	stop, err := http.NewRequest(http.MethodPost, "http://"+addr+"/"+server.Token()+"/api/stop", nil)
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
