package middleware

import "net/http"

// allowedOrigins lists the browser origins permitted to make credentialed
// (cookie-carrying) requests. A wildcard "*" cannot be combined with
// Access-Control-Allow-Credentials, so origins must be named explicitly.
var allowedOrigins = map[string]bool{
	"http://localhost:5173": true, // vite dev
	"http://localhost:4173": true, // vite preview
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// Responses differ per Origin, so caches must key on it.
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
