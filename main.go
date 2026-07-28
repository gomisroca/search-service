package main

import (
	"log"
	"net/http"
	"search-service/internal/config"
	"search-service/internal/handlers"
	"search-service/internal/index"
	"search-service/internal/middleware"
	"time"
)

func main() {
	cfg := config.Load()
	idx := index.New(cfg.MinTokenLength)

	if cfg.APIKey == "" {
		log.Println("WARNING: API_KEY is not set - all endpoints are open.")
	}

	mux := http.NewServeMux()

	mux.Handle("GET /health", &handlers.HealthHandler{Index: idx})
	mux.Handle("POST /documents",
		middleware.RequireAPIKey(cfg.APIKey, &handlers.DocumentsHandler{
			Index: idx, MaxDocuments: 
			cfg.MaxDocuments,
		}))
	mux.Handle("GET /documents/{id}",
		middleware.RequireAPIKey(cfg.APIKey, &handlers.DocumentHandler{Index: idx}))
	mux.Handle("DELETE /documents/{id}",
		middleware.RequireAPIKey(cfg.APIKey, &handlers.DocumentHandler{Index: idx}))
	mux.Handle("GET /search",
		middleware.RequireAPIKey(cfg.APIKey, &handlers.SearchHandler{
			Index: idx,
			DefaultSearchLimit: cfg.DefaultSearchLimit,
			MaxSearchLimit: cfg.MaxSearchLimit,
		}))
	
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler:  withCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Search service listening on :%s (min_token_length=%d)", cfg.Port, cfg.MinTokenLength)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
