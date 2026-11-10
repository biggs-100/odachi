package probe

// The seeded arm: the same small service with four defects planted in it, one
// for each lens. The manifest names them by line; this file is the exam.

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type server struct {
	db *sql.DB
}

type user struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (s *server) user(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

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
