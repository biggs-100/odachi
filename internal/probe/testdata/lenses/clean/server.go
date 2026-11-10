package probe

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

	var u user
	err := s.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}
