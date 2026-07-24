package models

import "time"

// body for POST /documents.
type IndexRequest struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
 
// stored representation of an indexed document.
type Document struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	IndexedAt time.Time      `json:"indexed_at"`
}
 
// one hit in a search response.
type SearchResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
}
 
// body returned from GET /search.
type SearchResponse struct {
	Query   string         `json:"query"`
	Total   int            `json:"total"`
	Results []SearchResult `json:"results"`
}
 
// returned after a successful POST /documents.
type IndexResponse struct {
	ID        string    `json:"id"`
	IndexedAt time.Time `json:"indexed_at"`
}
 
