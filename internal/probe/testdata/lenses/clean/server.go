package probe

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// queryTimeout bounds the database call so a slow query cannot hold the request.
const queryTimeout = 2 * time.Second

type server struct {
	db *sql.DB
}

type user struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// callerKey is the context key the authentication middleware uses to pass the
// authenticated caller's id to handlers.
type callerKey struct{}

// withCaller returns a context that carries id as the authenticated caller.
func withCaller(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, callerKey{}, id)
}

// callerID returns the authenticated caller's id and whether the request
// carries one.
func callerID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(callerKey{}).(int64)
	return id, ok
}

// user serves the authenticated caller's own record.
func (s *server) user(w http.ResponseWriter, r *http.Request) {
	caller, ok := callerID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	var u user
	err := s.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", caller).Scan(&u.ID, &u.Name)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("%s %s: query user: %v", r.Method, r.URL.Path, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("%s %s: write user: %v", r.Method, r.URL.Path, err)
	}
}
