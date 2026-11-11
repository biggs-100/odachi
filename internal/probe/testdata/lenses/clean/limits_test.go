package probe

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitCapsAtTheMaximum(t *testing.T) {
	if got := limit(5000); got != maxSize {
		t.Fatalf("limit(5000) = %d, want %d", got, maxSize)
	}
}

func TestLimitKeepsASmallerValue(t *testing.T) {
	if got := limit(10); got != 10 {
		t.Fatalf("limit(10) = %d, want 10", got)
	}
}

func TestLimitTreatsANegativeRequestAsZero(t *testing.T) {
	if got := limit(-1); got != 0 {
		t.Fatalf("limit(-1) = %d, want 0", got)
	}
}

func TestNewCacheSetsATimeout(t *testing.T) {
	if got := newCache().client.Timeout; got <= 0 {
		t.Fatalf("newCache timeout = %v, want a positive bound", got)
	}
}

func TestCacheSetAndGetRoundTrip(t *testing.T) {
	var c cache
	c.set("key", "value")

	got, ok := c.get("key")
	if !ok {
		t.Fatal("get missed a key that was set")
	}
	if got != "value" {
		t.Fatalf("get returned %q, want %q", got, "value")
	}
}

func TestCacheGetMissesAnUnknownKey(t *testing.T) {
	var c cache
	if _, ok := c.get("missing"); ok {
		t.Fatal("get reported a value for a key that was never set")
	}
}

func TestUserRejectsARequestWithoutACaller(t *testing.T) {
	s := &server{}
	rec := httptest.NewRecorder()

	s.user(rec, httptest.NewRequest(http.MethodGet, "/user", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUserReturnsNotFoundWhenNoRowMatches(t *testing.T) {
	s := &server{db: openTestDB(t, &fakeDriver{})}
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req = req.WithContext(withCaller(req.Context(), 7))
	rec := httptest.NewRecorder()

	s.user(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUserReturnsServerErrorWhenTheQueryFails(t *testing.T) {
	s := &server{db: openTestDB(t, &fakeDriver{err: errors.New("connection refused")})}
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req = req.WithContext(withCaller(req.Context(), 7))
	rec := httptest.NewRecorder()

	s.user(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUserReadsTheCallerAndIgnoresTheRequestedID(t *testing.T) {
	d := &fakeDriver{row: &user{ID: 7, Name: "Ada"}}
	s := &server{db: openTestDB(t, d)}
	req := httptest.NewRequest(http.MethodGet, "/user?id=99", nil)
	req = req.WithContext(withCaller(req.Context(), 7))

	s.user(httptest.NewRecorder(), req)

	if len(d.args) != 1 || d.args[0].Value != int64(7) {
		t.Fatalf("query args = %v, want the caller id 7", d.args)
	}
}

func TestUserWritesTheUserAsJSON(t *testing.T) {
	d := &fakeDriver{row: &user{ID: 7, Name: "Ada"}}
	s := &server{db: openTestDB(t, d)}
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req = req.WithContext(withCaller(req.Context(), 7))
	rec := httptest.NewRecorder()

	s.user(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var got user
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode the body: %v", err)
	}
	if got != (user{ID: 7, Name: "Ada"}) {
		t.Fatalf("body = %+v, want the caller's user", got)
	}
}

func TestUserDoesNotWriteASecondResponseAfterTheBodyWriteFails(t *testing.T) {
	d := &fakeDriver{row: &user{ID: 7, Name: "Ada"}}
	s := &server{db: openTestDB(t, d)}
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req = req.WithContext(withCaller(req.Context(), 7))
	w := &failingWriter{header: make(http.Header)}

	s.user(w, req)

	if w.writes != 1 {
		t.Fatalf("handler wrote %d times, want one failed write", w.writes)
	}
	if w.status != 0 {
		t.Fatalf("handler wrote status %d after the body write failed", w.status)
	}
}

type failingWriter struct {
	header http.Header
	status int
	writes int
}

func (w *failingWriter) Header() http.Header { return w.header }

func (w *failingWriter) WriteHeader(status int) { w.status = status }

func (w *failingWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, errors.New("connection closed")
}

func TestCacheFetchReturnsTheBody(t *testing.T) {
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusOK, strings.NewReader("hello")), nil
	}))

	body, err := c.fetch(context.Background(), "https://example.test/data")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if body != "hello" {
		t.Fatalf("fetch returned %q, want %q", body, "hello")
	}
}

func TestCacheFetchRejectsANonSuccessStatus(t *testing.T) {
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusServiceUnavailable, strings.NewReader("down")), nil
	}))

	if _, err := c.fetch(context.Background(), "https://example.test/data"); err == nil {
		t.Fatal("fetch accepted a 503 response")
	}
}

func TestCacheFetchReturnsAReadError(t *testing.T) {
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusOK, errorReader{}), nil
	}))

	if _, err := c.fetch(context.Background(), "https://example.test/data"); err == nil {
		t.Fatal("fetch returned a truncated body as a success")
	}
}

func TestCacheFetchRejectsABodyOverTheLimit(t *testing.T) {
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusOK, bytes.NewReader(make([]byte, maxResponseSize+1))), nil
	}))

	if _, err := c.fetch(context.Background(), "https://example.test/data"); err == nil {
		t.Fatal("fetch accepted a body over the limit")
	}
}

func TestCacheFetchGivesUpAfterTheLastAttempt(t *testing.T) {
	calls := 0
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("connection refused")
	}))

	if _, err := c.fetch(context.Background(), "https://example.test/data"); err == nil {
		t.Fatal("fetch returned nil after every attempt failed")
	}
	if calls != attempts {
		t.Fatalf("fetch made %d attempts, want %d", calls, attempts)
	}
}

func TestCacheFetchReturnsARequestBuildError(t *testing.T) {
	c := newTestCache(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("the transport ran for a request that could not be built")
		return nil, nil
	}))

	if _, err := c.fetch(context.Background(), "://"); err == nil {
		t.Fatal("fetch accepted an invalid URL")
	}
}

// fakeDriver answers the handler's user query from memory.
type fakeDriver struct {
	row  *user
	err  error
	args []driver.NamedValue
}

func (d *fakeDriver) Open(string) (driver.Conn, error) {
	return &fakeConn{driver: d}, nil
}

type fakeConn struct {
	driver *fakeDriver
}

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported")
}

func (c *fakeConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
	c.driver.args = append([]driver.NamedValue(nil), args...)
	if c.driver.err != nil {
		return nil, c.driver.err
	}
	return &fakeRows{row: c.driver.row}, nil
}

type fakeRows struct {
	row  *user
	sent bool
}

func (r *fakeRows) Columns() []string { return []string{"id", "name"} }

func (r *fakeRows) Close() error { return nil }

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.sent || r.row == nil {
		return io.EOF
	}
	r.sent = true
	dest[0] = r.row.ID
	dest[1] = r.row.Name
	return nil
}

type fakeConnector struct {
	driver *fakeDriver
}

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeConn{driver: c.driver}, nil
}

func (c fakeConnector) Driver() driver.Driver {
	return c.driver
}

func openTestDB(t *testing.T, d *fakeDriver) *sql.DB {
	t.Helper()
	db := sql.OpenDB(fakeConnector{driver: d})
	t.Cleanup(func() { db.Close() })
	return db
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestCache(t *testing.T, transport http.RoundTripper) *cache {
	t.Helper()
	c := newCache()
	c.client = &http.Client{Transport: transport}
	return c
}

func testResponse(status int, body io.Reader) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     make(http.Header),
		Body:       io.NopCloser(body),
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("connection reset")
}
