package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"newsletter/internal/infrastructure/workerpool"
	"newsletter/internal/infrastructure/workerpool/jobs"
	notifications "newsletter/internal/notifications/domain"
	"newsletter/internal/subscriptions/domain"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SubscriptionHandler struct {
	ss domain.SubscriptionService
	es notifications.EmailService
	wp *workerpool.WorkerPool
}

func NewSubscriptionHandler(ss domain.SubscriptionService, es notifications.EmailService, wp *workerpool.WorkerPool) *SubscriptionHandler {
	return &SubscriptionHandler{ss: ss, es: es, wp: wp}
}

// SubscribeRequest represents the payload for subscribing to a newsletter.
type SubscribeRequest struct {
	Email string `json:"email"` // Email of the subscriber
}

// SubscribeResponse represents the response returned after a subscription is created.

type SubscribeResponse struct {
	ID           string    `json:"id"`
	NewsletterID uuid.UUID `json:"newsletter_id"`
	Email        string    `json:"email"`
	Confirmed    bool      `json:"confirmed"`
	CreatedAt    time.Time `json:"created_at"`
}

func (sh *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	// Retrieve newsletter ID from path parameters
	vars := mux.Vars(r)
	newsletterIDStr, found := vars["newsletter_id"]
	if !found {
		http.Error(w, "newsletter ID is missing from path parameters", http.StatusBadRequest)
		return
	}
	// Convert string to uuid.UUID
	newsletterID, err := uuid.Parse(newsletterIDStr)
	if err != nil {
		slog.Warn("invalid newsletter ID", "newsletterID", newsletterIDStr, "error", err)
		http.Error(w, "invalid newsletter ID format", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	subscription := domain.Subscription{
		NewsletterID: newsletterID,
		Email:        request.Email,
	}
	newSubscription, err := sh.ss.Subscribe(&subscription)
	if err != nil {
		http.Error(w, "failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send confirmation email to the subscriber
	job := jobs.SendEmailJob{
		Email: notifications.Email{
			To:      newSubscription.Email,
			Subject: "Confirmation",
			Text:    "",
			HTML:    "C",
		},
		Service: sh.es,
	}
	sh.wp.Submit(&job)

	// Respond with created subscription in JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	subscribeResponse := SubscribeResponse{
		ID:           newSubscription.ID,
		NewsletterID: newSubscription.NewsletterID,
		Email:        newSubscription.Email,
		Confirmed:    newSubscription.Confirmed,
		CreatedAt:    newSubscription.CreatedAt,
	}
	if err := json.NewEncoder(w).Encode(subscribeResponse); err != nil {
		slog.Error("failed to encode subscription response",
			"newsletter_id", newSubscription.NewsletterID,
			"email", newSubscription.Email,
			"error", err,
		)
	}
}

func (sh *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {

}
