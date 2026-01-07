package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Subscription represents a newsletter subscription.
type Subscription struct {
	ID           string    `json:"id"`            // Unique ID of the subscription
	NewsletterID uuid.UUID `json:"newsletter_id"` // ID of the newsletter this subscription belongs to
	Email        string    `json:"email"`         // Email of the subscriber
	CreatedAt    time.Time `json:"created_at"`    // Creation time of the subscription
}

// SubscriptionRepository is an interface that contains a collection of method signatures
// which will be implemented in persistence level.
type SubscriptionRepository interface {
	// Subscribe adds a new subscription for a newsletter
	Subscribe(ctx context.Context, newsletterID uuid.UUID, email string) (*Subscription, error)

	// Unsubscribe removes a subscription
	Unsubscribe(ctx context.Context, subscriptionID string) error
}
