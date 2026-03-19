package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// SwaggerBasicAuth protects Swagger UI HTML pages with HTTP Basic Auth
// while allowing static assets (.js, .css, .json, .png) through without auth.
func SwaggerBasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Allow static assets through without auth
			if strings.HasSuffix(path, ".js") ||
				strings.HasSuffix(path, ".css") ||
				strings.HasSuffix(path, ".json") ||
				strings.HasSuffix(path, ".yaml") ||
				strings.HasSuffix(path, ".png") ||
				strings.HasSuffix(path, ".ico") {
				next.ServeHTTP(w, r)
				return
			}

			// Require basic auth for HTML pages
			user, pass, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger UI"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
