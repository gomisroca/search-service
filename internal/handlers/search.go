package handlers

import (
	"net/http"
	"search-service/internal/index"
	"search-service/internal/models"
	"strconv"
	"strings"
)

type SearchHandler struct {
	Index 				*index.Index
	DefaultSearchLimit 	int
	MaxSearchLimit 		int
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Only GET requests are supported")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "Missing q query parameter")
		return
	}

	limit := h.DefaultSearchLimit
	if ls := r.URL.Query().Get("limit"); ls != "" {
		n, err := strconv.Atoi(ls)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "Limit must be a positive integer")
			return
		}
		limit = n
	}
	if limit > h.MaxSearchLimit {
		limit = h.MaxSearchLimit
	}

	results := h.Index.Search(query, limit)
	if results == nil {
		results = []models.SearchResult{}
	}

	writeJSON(w, http.StatusOK, models.SearchResponse{
		Query: query,
		Total: len(results),
		Results: results,
	})
}