package handlers

import (
	"fmt"
	"net/http"

	"search-service/internal/index"
)

type HealthHandler struct {
	Index *index.Index
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"documents": fmt.Sprintf("%d", h.Index.Len()),
	})
}