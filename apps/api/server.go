package main

import (
	"log"
	"net/http"
	"time"
)

// NewServer wires the mux and sensible timeouts.
func NewServer(addr string) *http.Server {
	mux := http.NewServeMux()

	// Routes (mount API under /)
	mux.HandleFunc("GET /health", HealthHandler)
	mux.HandleFunc("GET /ollama/health", OllamaHealthHandler) // test Ollama connection
	mux.HandleFunc("POST /ollama/test", OllamaTestHandler)    // test text generation
	mux.HandleFunc("GET /threads", ThreadsListHandler)       // stub
	mux.HandleFunc("GET /threads/{id}", ThreadDetailHandler) // stub
	mux.HandleFunc("POST /search", SearchHandler)            // stub
	mux.HandleFunc("POST /chat", ChatHandler)                // stub

	// (Later) Gmail OAuth:
	// mux.HandleFunc("POST /auth/google/start", GoogleAuthStart)
	// mux.HandleFunc("GET /auth/google/callback", GoogleAuthCallback)

	// Wrap with middleware chain.
	handler := withCORS(withLogging(mux))

	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 5 * time.Minute, // Match Ollama client timeout
		IdleTimeout:  6 * time.Minute, // Slightly longer than write timeout
	}
}

// --- Middleware (stdlib) ---

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logResponseWriter{ResponseWriter: w, code: 200}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, lw.code, time.Since(start))
	})
}

type logResponseWriter struct {
	http.ResponseWriter
	code int
}

func (lw *logResponseWriter) WriteHeader(statusCode int) {
	lw.code = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust origin as needed during dev
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
