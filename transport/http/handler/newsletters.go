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

// NewsletterHandler handles HTTP requests related to newsletters,
// including creation and retrieval.
type NewsletterHandler struct {
	ns domain.NewsletterService
}

// NewNewsletterHandler creates a new NewsletterHandler.
func NewNewsletterHandler(ns domain.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{ns: ns}
}

// Create handles creating a new newsletter.
//
// Route:
//
//	POST /newsletters
//
// Description:
//
//	Creates a new newsletter owned by the authenticated user. The owner ID
//	is extracted from the request context (set by authentication middleware).
//
// Request Body (application/json):
//
//	{
//	  "name": "My Newsletter",
//	  "description": "Weekly updates about tech"
//	}
//
// Responses:
//
//	201 Created
//	  {
//	    "id": "uuid",
//	    "name": "My Newsletter",
//	    "description": "Weekly updates about tech",
//	    "owner_id": "uuid",
//	    "created_at": "2026-01-10T12:00:00Z"
//	  }
//
// Responses:
//
//	201 Created
//	  {
//	    "id": "uuid",
//	    "name": "My Newsletter",
//	    "description": "Weekly updates about tech",
//	    "owner_id": "uuid",
//	    "created_at": "2026-01-10T12:00:00Z"
//	  }
//
//	400 Bad Request
//	  - Invalid JSON body
//	  - Invalid owner ID
//
//	401 Unauthorized
//	  - Missing or invalid authentication context
//
//	500 Internal Server Error
//	  - Newsletter creation failure
//
// Side Effects:
//   - Persists a new newsletter owned by the authenticated user
func (nh *NewsletterHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var newsletter domain.Newsletter
	if err := json.NewDecoder(r.Body).Decode(&newsletter); err != nil {
		slog.Warn("failed to decode request body", "error", err)
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	newsletter.OwnerID = ownerID

	newNewsletter, err := nh.ns.Create(&newsletter)
	if err != nil {
		slog.Error("failed to create newsletter", "owner_id", newsletter.OwnerID, "name", newsletter.Name, "error", err)
		http.Error(w, "failed to create newsletter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newNewsletter); err != nil {
		slog.Error("failed to encode newsletter response", "owner_id", ownerID, "error", err)
	}
}

// GetAll handles retrieving all newsletters for the authenticated user.
//
// Route:
//
//	GET /newsletters
//
// Description:
//
//	Returns a paginated list of newsletters owned by the authenticated user.
//	Pagination is controlled via optional query parameters.
//
// Query Parameters:
//
//	limit (int, optional) - Number of newsletters per page (default: 10)
//	page  (int, optional) - Page number (default: 1)
//
// Responses:
//
//	200 OK
//	  [
//	    {
//	      "id": "uuid",
//	      "name": "My Newsletter",
//	      "description": "Weekly updates about tech",
//	      "owner_id": "uuid",
//	      "created_at": "2026-01-10T12:00:00Z"
//	    }
//	  ]
//
//	400 Bad Request
//	  - Invalid owner ID
//
//	401 Unauthorized
//	  - Missing or invalid authentication context
//
//	500 Internal Server Error
//	  - Newsletter retrieval failure
//
// Side Effects:
//   - None
func (nh *NewsletterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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
		limit = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
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
