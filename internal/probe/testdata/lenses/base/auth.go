package probe

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

// authKey is the context key the middleware uses to pass the authenticated
// caller's id to handlers.
type authKey struct{}

// withCaller returns a context that carries id as the authenticated caller.
func withCaller(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, authKey{}, id)
}

// callerID returns the authenticated caller's id and whether the request
// carries one.
func callerID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(authKey{}).(int64)
	return id, ok
}

// authenticate installs the caller in the context of every request it lets
// through, and answers 401 for a request it cannot resolve.
func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := sessionCaller(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(withCaller(r.Context(), id)))
	})
}

// sessionCaller reads the caller the session names.
func sessionCaller(r *http.Request) (int64, error) {
	raw := r.Header.Get("X-Caller-Id")
	if raw == "" {
		return 0, errors.New("no session")
	}
	return strconv.ParseInt(raw, 10, 64)
}
