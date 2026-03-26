package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github/yeshu2004/go-epics/crawler"
	db "github/yeshu2004/go-epics/db"
)

type Server struct {
	PostgresDB *db.PostgresDB
	Crawler    *crawler.Crawler
	mu         sync.Mutex
}

type CrawlRequest struct {
	SeedURLs []string `json:"seed_urls"`
}

type CrawlResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type StatsResponse struct {
	Crawled    int64 `json:"crawled"`
	Duplicates int64 `json:"duplicates"`
}

func NewServer(postgresDB *db.PostgresDB, crawler *crawler.Crawler) *Server {
	return &Server{
		PostgresDB: postgresDB,
		Crawler:    crawler,
	}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/crawl/start", s.enableCORS(s.handleCrawlStart))
	mux.HandleFunc("/api/search", s.enableCORS(s.handleSearch))
	mux.HandleFunc("/api/stats", s.enableCORS(s.handleStats))
	mux.HandleFunc("/api/domains", s.enableCORS(s.handleDomains))

	log.Printf("API Server starting on port %s", port)
	return http.ListenAndServe(":"+port, mux)
}

func (s *Server) enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleCrawlStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.SeedURLs) == 0 {
		http.Error(w, "seed_urls cannot be empty", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	go s.Crawler.Start(req.SeedURLs)

	resp := CrawlResponse{
		Status:  "started",
		Message: "Crawling started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	results, err := s.PostgresDB.SearchWords(context.Background(), query)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	crawled, duplicates := s.Crawler.GetStats()

	resp := StatsResponse{
		Crawled:    crawled,
		Duplicates: duplicates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.PostgresDB.GetDomainStats(context.Background())
	if err != nil {
		http.Error(w, "Failed to get domain stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
