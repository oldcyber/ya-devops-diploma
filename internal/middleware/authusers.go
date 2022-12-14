package middleware

import (
	"net/http"

	"github.com/oldcyber/ya-devops-diploma/internal/auth"
)

// SetMiddlewareAuthentication set the authentication
func SetMiddlewareAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth.TokenValid(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Set("uid", uid)
		next(w, r)
	}
}
