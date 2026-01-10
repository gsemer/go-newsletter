package domain

import (
	"context"
	"time"
)

// Subscription represents a newsletter subscription.
type Subscription struct {
	ID               string    `firestore:"-" json:"id"`                       // Firestore document ID
	NewsletterID     string    `firestore:"newsletterId" json:"newsletter_id"` // Newsletter ID
	Email            string    `firestore:"email" json:"email"`                // Email of the subscriber
	UnsubscribeToken string    `firestore:"unsubscribeToken" json:"-"`         // Token to unsubscribe
	CreatedAt        time.Time `firestore:"createdAt" json:"created_at"`       // Creation time
}

// SubscriptionService is an interface that contains a collection of method signatures
// which will be implemented in application level.
type SubscriptionService interface {
	// Subscribe adds a new subscription for a newsletter
	Subscribe(subscription *Subscription) (*Subscription, error)

	// Unsubscribe removes a subscription
	Unsubscribe(unsubscribeToken string) error
}

// SubscriptionRepository is an interface that contains a collection of method signatures
// which will be implemented in persistence level.
type SubscriptionRepository interface {
	Subscribe(ctx context.Context, subscription *Subscription) (*Subscription, error)
	Unsubscribe(ctx context.Context, unsubscribeToken string) error
}
