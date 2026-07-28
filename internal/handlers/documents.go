package handlers

import (
	"encoding/json"
	"net/http"
	"search-service/internal/index"
	"search-service/internal/models"
	"strings"
	"time"
)

type DocumentsHandler struct {
	Index *index.Index
	MaxDocuments int
}

func (h *DocumentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleIndex(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Only POST requests are supported for /documents")
	}
}

func (h *DocumentsHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	var req models.IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.Content = strings.TrimSpace(req.Content)

	if req.ID == "" {
		writeError(w, http.StatusUnprocessableEntity, "ID is required")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusUnprocessableEntity, "Content is required")
		return
	}

	if h.MaxDocuments > 0 && h.Index.Len() >= h.MaxDocuments {
		if _, exists := h.Index.Get(req.ID); !exists {
			writeError(w, http.StatusServiceUnavailable, "Index is full")
			return
		}
	}

	now := time.Now().UTC()
	doc := models.Document{
		ID: req.ID,
		Content: req.Content,
		Metadata: req.Metadata,
		IndexedAt: now,
	}
	h.Index.Add(doc)

	writeJSON(w, http.StatusCreated, models.IndexResponse{
		ID: req.ID,
		IndexedAt: now,
	})
}

type DocumentHandler struct {
	Index *index.Index
}


func (h *DocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	switch r.Method {
	case http.MethodGet:
		doc, ok := h.Index.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "Document not found")
			return
		}
		writeJSON(w, http.StatusOK, doc)
	case http.MethodDelete:
		if !h.Index.Delete(id) {
			writeError(w, http.StatusNotFound, "Document not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Only GET and DELETE are supported for /documents/{id}")
	}
}
