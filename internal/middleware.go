package internal

import (
	// "log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func applyMiddleWare(r *chi.Mux) {
	
	r.Use(middleware.Logger) // add log middleware
	r.Use(limitUploadSize) // add limit middleware
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