package probe

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type server struct {
	db *sql.DB
}

type user struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type callerKey struct{}

func (s *server) user(w http.ResponseWriter, r *http.Request) {
	caller, ok := r.Context().Value(callerKey{}).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	id := r.URL.Query().Get("id")
	if id == "" {
		id = strconv.FormatInt(caller, 10)
	}
	query := "SELECT id, name FROM users WHERE id = " + id
	var u user
	err := s.db.QueryRowContext(ctx, query).Scan(&u.ID, &u.Name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}
