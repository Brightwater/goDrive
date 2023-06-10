package internal

import (
	// "log"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const maxRequestBodySize int64 = 2e+10 // 20Gb

func applyMiddleWare(r *chi.Mux) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	})

	r.Use(cors.Handler)
	r.Use(middleware.Logger) // add log middleware
	r.Use(limitUploadSize)   // add limit middleware

	// Handle OPTIONS requests
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, Access-Control-Allow-Origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		w.WriteHeader(http.StatusNoContent)
	})

	// Serve the index.html file as the root path
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join("static", "index.html"))
	})

	// Serve static files from the "public" directory
	r.Handle("/_app/*", http.FileServer(http.Dir("static")))
}

// func logRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("[%s] %s %s\n", r.RemoteAddr, r.Method, r.URL)
// 		next.ServeHTTP(w, r)
// 	})
// }

func limitUploadSize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		next.ServeHTTP(w, r)
	})
}
