package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"newsletter/config"
	"newsletter/internal/infrastructure/workerpool"
	"newsletter/internal/infrastructure/workerpool/jobs"
	notifications "newsletter/internal/notifications/domain"
	"newsletter/internal/subscriptions/domain"
	"time"

	"github.com/gorilla/mux"
)

type SubscriptionHandler struct {
	ss domain.SubscriptionService
	es notifications.EmailService
	wp workerpool.JobSubmiter
}

func NewSubscriptionHandler(ss domain.SubscriptionService, es notifications.EmailService, wp workerpool.JobSubmiter) *SubscriptionHandler {
	return &SubscriptionHandler{ss: ss, es: es, wp: wp}
}

// SubscribeRequest represents the payload for subscribing to a newsletter.
type SubscribeRequest struct {
	Email string `json:"email"` // Email of the subscriber
}

// SubscribeResponse represents the response returned after a subscription is created.
type SubscribeResponse struct {
	ID           string    `json:"id"`
	NewsletterID string    `json:"newsletter_id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

// Subscribe handles newsletter subscription requests.
//
// Route:
//
//	POST /subscriptions/{newsletter_id}
//
// Description:
//
//	Subscribes an email address to a specific newsletter. Upon successful
//	subscription, a confirmation email is sent containing an unsubscribe link.
//
// Path Parameters:
//
//	newsletter_id (string) - The ID of the newsletter to subscribe to
//
// Request Body (application/json):
//
//	{
//	  "email": "user@example.com"
//	}
//
// Responses:
//
//	201 Created
//	  {
//	    "id": "subscription_id",
//	    "newsletter_id": "newsletter_id",
//	    "email": "user@example.com",
//	    "created_at": "2026-01-10T12:00:00Z"
//	  }
//
//	400 Bad Request
//	  - Missing newsletter_id in path
//	  - Invalid JSON body
//
//	500 Internal Server Error
//	  - Subscription creation failure
//
// Side Effects:
//   - Sends a confirmation email containing an unsubscribe link with a token.
func (sh *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	newsletterID, found := vars["newsletter_id"]
	if !found {
		http.Error(w, "newsletter ID is missing from path parameters", http.StatusBadRequest)
		return
	}

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

	// Send confirmation email to the subscriber with unsubscribe link
	job := jobs.SendEmailJob{
		Email: notifications.Email{
			To:      newSubscription.Email,
			Subject: "Confirmation",
			Text: fmt.Sprintf(
				`You are receiving this email because you subscribed to this newsletter.
                If you no longer wish to receive these emails, you can unsubscribe using the link below:
                %s/subscriptions/unsubscribe?token=%s`,
				config.GetEnv("BASE_URL", ""),
				newSubscription.UnsubscribeToken,
			),
			HTML: fmt.Sprintf(
				`<p>You are receiving this email because you subscribed to this newsletter.</p>
				<p>If you no longer wish to receive these emails, you can
				<a href="%s/subscriptions/unsubscribe?token=%s">unsubscribe here</a>.</p>`,
				config.GetEnv("BASE_URL", ""),
				newSubscription.UnsubscribeToken,
			),
		},
		Service: sh.es,
	}
	sh.wp.Submit(&job)

	// Immediate response with created subscription in JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	subscribeResponse := SubscribeResponse{
		ID:           newSubscription.ID,
		NewsletterID: newSubscription.NewsletterID,
		Email:        newSubscription.Email,
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

// Unsubscribe removes a subscription using an unsubscribe token.
//
// This endpoint allows a user to unsubscribe from a newsletter by providing
// a unique token, typically included in the newsletter email. If the token
// is valid, the associated subscription is deleted from the system.
//
// HTTP Method: DELETE
//
// Query Parameters:
//   - token (string) - The unique unsubscribe token identifying the subscription.
//
// Behavior:
//   - Returns 400 Bad Request if the token is missing.
//   - Returns 404 Not Found if no subscription matches the given token.
//   - Returns 204 No Content on successful unsubscription.
//
// Example usage:
//
//	DELETE /subscriptions/unsubscribe?token=abcd1234
//
// Notes:
//   - The unsubscribe token should be globally unique for each subscription.
func (sh *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	err := sh.ss.Unsubscribe(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
