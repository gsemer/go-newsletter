package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"newsletter/internal/newsletters/domain"
	userdomain "newsletter/internal/users/domain"
	"strconv"

	"github.com/google/uuid"
)

type NewsletterHandler struct {
	ns domain.NewsletterService
}

func NewNewsletterHandler(ns domain.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{ns: ns}
}

// Create handles creating a new newsletter.
// It expects JSON in the request body with fields "name" and "description".
// The owner ID is extracted from the request context.
// Returns the created newsletter in JSON format.
func (nh *NewsletterHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extract owner ID from context
	value := r.Context().Value(userdomain.UserID)
	ownerIDStr, ok := value.(string)
	if !ok {
		slog.Warn("owner ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		slog.Warn("invalid owner ID", "ownerID", ownerIDStr, "error", err)
		http.Error(w, "invalid identification", http.StatusBadRequest)
		return
	}

	// Decode request body into newsletter
	var newsletter domain.Newsletter
	if err := json.NewDecoder(r.Body).Decode(&newsletter); err != nil {
		slog.Warn("failed to decode request body", "error", err)
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set owner ID from context
	newsletter.OwnerID = ownerID

	// Call the service to create the newsletter
	newNewsletter, err := nh.ns.Create(&newsletter)
	if err != nil {
		slog.Error("failed to create newsletter", "owner_id", newsletter.OwnerID, "name", newsletter.Name, "error", err)
		http.Error(w, "failed to create newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with created newsletter in JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newNewsletter); err != nil {
		slog.Error("failed to encode newsletter response", "owner_id", ownerID, "error", err)
	}
}

// GetAll handles the request to list all newsletters for the authenticated user.
func (nh *NewsletterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Extract owner ID from context
	value := r.Context().Value(userdomain.UserID)
	ownerIDStr, ok := value.(string)
	if !ok {
		slog.Warn("owner ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		slog.Warn("invalid owner ID", "ownerID", ownerIDStr, "error", err)
		http.Error(w, "invalid identification", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10 // Default to 10 items
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1 // Default to first page
	}

	newsletters, err := nh.ns.GetAll(ownerID, limit, page)
	if err != nil {
		slog.Error("service failure during newsletter retrieval", "owner_id", ownerID, "error", err)
		http.Error(w, "failed to retrieve newsletters: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(newsletters); err != nil {
		slog.Error("failed to encode newsletters response", "owner_id", ownerID, "error", err)
	}
}
