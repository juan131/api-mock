package authn

import (
	"net/http"
	"strings"
)

const authenticateHeader string = `Bearer realm="example", error="invalid_token", error_description="invalid access token"`

// ApiKeyAuth implements a simple middleware handler for adding
// authentication based on an API key set in the X-API-KEY header.
func ApiKeyAuth(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-KEY") != key {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BearerTokenAuth implements a simple middleware handler for adding
// bearer http auth based on tokens to a route.
func BearerTokenAuth(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if strings.TrimPrefix(authHeader, "Bearer ") != token {
				w.Header().Add("WWW-Authenticate", authenticateHeader)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
