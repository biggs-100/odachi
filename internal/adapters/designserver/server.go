// Package designserver is the canvas: a local HTTP surface over the design folder
// that a person can open in a browser.
//
// It is an adapter and nothing else. It reads the same files the command reads,
// writes the same comments file, and serves the artboards as they are: the folder
// stays the protocol, so a canvas that is closed loses nothing and an agent that
// never opens it loses nothing either.
package designserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	domain "github.com/biggs-100/odachi/internal/domain/design"
)

// Options is one canvas: the project it serves, the design it opens first and the
// browser it may open.
type Options struct {
	// Dir is the project root: the server reads design/ under it.
	Dir string
	// Slug is the design the page opens first; empty opens the first one.
	Slug string
	// Canvas is the page served at /.
	Canvas []byte
}

// Server is a canvas over one project. One server serves one project: a second
// project is a second process, because the folder is the address.
type Server struct {
	opts Options

	mu      sync.Mutex
	clients map[chan string]bool
	last    map[string]time.Time
	// writeMu serialises the comment writes: two clicks at once must not read,
	// append and replace the same file in an order nobody chose.
	writeMu sync.Mutex

	http *http.Server
	// host is the address the canvas listens on: the origin a request must carry
	// to be its own. The Host header is the caller's, so it cannot decide.
	host string
	done chan struct{}
	once sync.Once
}

// New builds the canvas. It reads nothing until it is asked to.
func New(opts Options) *Server {
	return &Server{opts: opts, clients: map[chan string]bool{}, last: map[string]time.Time{}, done: make(chan struct{})}
}

// Done is closed when the canvas stops, so `open` can block for as long as a person
// is looking at it and return the moment somebody stops it.
func (s *Server) Done() <-chan struct{} { return s.done }

// Listen starts the server on a loopback port and returns the address and the URL
// a person clicks. The port is the kernel's choice unless the caller names one.
func (s *Server) Listen(port int) (addr, url string, err error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return "", "", err
	}

	s.http = &http.Server{Handler: s.routes(), ReadHeaderTimeout: 5 * time.Second}
	addr = listener.Addr().String()
	s.host = addr
	url = "http://" + addr + "/"
	if slug := strings.TrimSpace(s.opts.Slug); slug != "" {
		url += "?design=" + slug
	}
	go func() { _ = s.http.Serve(listener) }()
	return addr, url, nil
}

// Close stops the server. It never waits for a person to leave: the state file is
// the address, and a canvas nobody can reach is not running.
func (s *Server) Close() error {
	s.once.Do(func() { close(s.done) })
	if s.http == nil {
		return nil
	}
	return s.http.Close()
}

// routes is the whole surface: the page, the artboards, the JSON the page reads and
// the stream that tells it to reload.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.page)
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/designs", s.designs)
	mux.HandleFunc("/api/design/", s.design)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/stop", s.stop)
	mux.HandleFunc("/artboard/", s.artboard)
	return mux
}

// page serves the canvas. It is one file with its script inline: no build, no CDN,
// and the same instinct the artboards follow.
func (s *Server) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(s.opts.Canvas)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST"})
		return
	}
	if !s.ownPage(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopping"})
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = s.Close()
	}()
}

// designs lists the designs the project keeps, with their artboards and comments.
func (s *Server) designs(w http.ResponseWriter, r *http.Request) {
	all, err := s.readDesigns("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"designs": all})
}

// design returns one design, or adds and resolves its comments.
func (s *Server) design(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/design/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" || !isSlug(parts[0]) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no design"})
		return
	}
	slug := parts[0]

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		all, err := s.readDesigns(slug)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if len(all) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no design " + slug})
			return
		}
		writeJSON(w, http.StatusOK, all[0])
	case len(parts) == 2 && parts[1] == "comments" && r.Method == http.MethodPost:
		s.addComment(w, r, slug)
	case len(parts) == 3 && parts[1] == "comments" && r.Method == http.MethodPost:
		s.resolveComment(w, r, slug, parts[2])
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "not here"})
	}
}

// addComment appends one comment to the file the agents read, and never overwrites
// one it cannot parse: a canvas that destroys a person's notes is worse than no
// canvas.
func (s *Server) addComment(w http.ResponseWriter, r *http.Request, slug string) {
	if !s.ownPage(w, r) || !s.jsonOnly(w, r) {
		return
	}

	var in struct {
		File       string `json:"file"`
		Selector   string `json:"selector"`
		Excerpt    string `json:"excerpt"`
		Comment    string `json:"comment"`
		Screenshot string `json:"screenshot"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "the comment is not JSON: " + err.Error()})
		return
	}
	if strings.TrimSpace(in.Comment) == "" || strings.TrimSpace(in.File) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "a comment needs its artboard and its text"})
		return
	}

	s.writeMu.Lock()

	comments, err := s.readComments(slug)
	if err != nil {
		s.writeMu.Unlock()
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	comment := domain.Comment{
		ID:         newID(),
		File:       in.File,
		Selector:   in.Selector,
		Excerpt:    in.Excerpt,
		Comment:    in.Comment,
		Screenshot: in.Screenshot,
		CreatedAt:  time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Status:     domain.StatusOpen,
	}
	comments.Comments = append(comments.Comments, comment)
	if err := s.writeComments(slug, comments); err != nil {
		s.writeMu.Unlock()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeMu.Unlock()
	s.broadcast("comments")
	writeJSON(w, http.StatusCreated, comment)
}

func (s *Server) resolveComment(w http.ResponseWriter, r *http.Request, slug, id string) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	comments, err := s.readComments(slug)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	grown, missing := comments.Resolve([]string{id}, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
	if len(missing) > 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no comment " + id})
		return
	}
	if err := s.writeComments(slug, grown); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.broadcast("comments")
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

// artboard serves one HTML file of one design. Only a file the manifest lists is
// served: the canvas is not a file browser over the project.
func (s *Server) artboard(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/artboard/")
	parts := strings.SplitN(strings.Trim(rest, "/"), "/", 2)
	if len(parts) != 2 || !isSlug(parts[0]) || parts[1] == "" || parts[1] != filepath.Base(parts[1]) {
		http.NotFound(w, r)
		return
	}

	all, err := s.readDesigns(parts[0])
	if err != nil || len(all) == 0 {
		http.NotFound(w, r)
		return
	}
	listed := false
	for _, artboard := range all[0].Manifest.Artboards {
		if artboard.File == parts[1] {
			listed = true
			break
		}
	}
	if !listed {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.opts.Dir, filepath.FromSlash(domain.Dir), parts[0], parts[1])
	raw, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(raw)
}

// events is the live reload: one line per change, and the page decides what to
// reload. It polls the folder instead of watching it, so the binary needs no
// dependency and the behaviour is the same on every OS.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 8)
	s.mu.Lock()
	s.clients[ch] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case what := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", what)
			flusher.Flush()
		case <-ticker.C:
			if changed := s.changed(); len(changed) > 0 {
				fmt.Fprintf(w, "data: %s\n\n", strings.Join(changed, ","))
				flusher.Flush()
			}
		}
	}
}

// changed reports what moved since the last look: one artboard, or the manifest and
// comments of a design. It keeps the opinion to itself and tells the page the name.
func (s *Server) changed() []string {
	root := filepath.Join(s.opts.Dir, filepath.FromSlash(domain.Dir))
	designs, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	moved := []string{}
	for _, entry := range designs {
		if !entry.IsDir() || !isSlug(entry.Name()) {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, err := file.Info()
			if err != nil {
				continue
			}
			key := entry.Name() + "/" + file.Name()
			if before, seen := s.last[key]; !seen {
				s.last[key] = info.ModTime()
				continue
			} else if info.ModTime() != before {
				s.last[key] = info.ModTime()
				// The comments file is the panel, not an artboard: the page reloads
				// one or the other, and telling it "design.json" would reload a frame.
				if file.Name() == domain.CommentsFile {
					moved = append(moved, "comments")
				} else {
					moved = append(moved, entry.Name()+"/"+file.Name())
				}
			}
		}
	}
	return moved
}

func (s *Server) broadcast(what string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- what:
		default:
		}
	}
}

// Design is one design as the page reads it: the manifest and the comments file.
type Design struct {
	Slug     string            `json:"slug"`
	Manifest domain.Manifest   `json:"manifest"`
	Comments domain.Comments   `json:"comments"`
	URL      map[string]string `json:"urls"`
}

func (s *Server) readDesigns(slug string) ([]Design, error) {
	root := filepath.Join(s.opts.Dir, filepath.FromSlash(domain.Dir))
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []Design{}, nil
		}
		return nil, err
	}

	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() && isSlug(entry.Name()) && (slug == "" || entry.Name() == slug) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	out := make([]Design, 0, len(names))
	for _, name := range names {
		dir := filepath.Join(root, name)
		manifestRaw, err := os.ReadFile(filepath.Join(dir, domain.ManifestFile))
		if err != nil {
			continue
		}
		manifest, err := domain.ParseManifest(manifestRaw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		comments, err := s.readComments(name)
		if err != nil {
			return nil, err
		}
		urls := make(map[string]string, len(manifest.Artboards))
		for _, artboard := range manifest.Artboards {
			urls[artboard.File] = "/artboard/" + name + "/" + artboard.File
		}
		out = append(out, Design{Slug: name, Manifest: manifest, Comments: comments, URL: urls})
	}
	return out, nil
}

func (s *Server) readComments(slug string) (domain.Comments, error) {
	path := filepath.Join(s.opts.Dir, filepath.FromSlash(domain.Dir), slug, domain.CommentsFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Comments{Version: domain.Version, Comments: []domain.Comment{}}, nil
		}
		return domain.Comments{}, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return domain.Comments{Version: domain.Version, Comments: []domain.Comment{}}, nil
	}
	return domain.ParseComments(raw)
}

// writeComments replaces the file atomically: a crash mid-write leaves the old one,
// never half of the new one.
func (s *Server) writeComments(slug string, comments domain.Comments) error {
	dir := filepath.Join(s.opts.Dir, filepath.FromSlash(domain.Dir), slug)
	path := filepath.Join(dir, domain.CommentsFile)
	data := domain.RenderComments(comments)

	temp, err := os.CreateTemp(dir, ".comments-*.json")
	if err != nil {
		return err
	}
	name := temp.Name()
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		os.Remove(name)
		return err
	}
	if err := temp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	return os.Rename(name, path)
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func isSlug(name string) bool {
	return name != "" && name == filepath.Base(name) && !strings.HasPrefix(name, ".")
}

// newID is the comment identity: eight hex characters, the same shape the skill
// documents, minted without a dependency.
func newID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xffffffff)
	}
	return hex.EncodeToString(buf)
}

// ownPage refuses a caller this canvas did not serve: a browser sends Origin on a
// cross-origin POST, and a page from the web must not append a comment or stop the
// canvas just because the port answers on loopback.
func (s *Server) ownPage(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	own := s.host
	if own == "" {
		own = r.Host
	}
	if err != nil || parsed.Host != own {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "cross-origin request refused: the canvas answers only its own page"})
		return false
	}
	return true
}

// jsonOnly refuses a body that is not JSON: a simple text/plain form post needs no
// preflight, and refusing it is what keeps a web page from writing through a canvas
// it never opened.
func (s *Server) jsonOnly(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "the canvas reads application/json only"})
		return false
	}
	return true
}
