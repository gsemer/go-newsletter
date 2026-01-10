package firebase

import (
	"context"
	"newsletter/internal/subscriptions/domain"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type SubscriptionRepository struct {
	db *firestore.Client
}

func NewSubscriptionRepository(db *firestore.Client) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Subscribe persists a new subscription in the database.
//
// Parameters:
//   - ctx: context for managing cancellation and timeouts
//   - subscription: pointer to a Subscription domain object containing
//     the newsletter ID and subscriber email. The ID and timestamps
//     will be populated by this method.
//
// Behavior:
//   - Generates a new unsubscribe token for the subscription.
//   - Sets the CreatedAt timestamp to the current time.
//   - Adds the subscription to the "subscriptions" collection in the database.
//   - Populates the subscription.ID field with the database-generated document ID.
//
// Returns:
//   - pointer to the created Subscription object with ID and unsubscribe token set
//   - error if the operation fails
func (sr *SubscriptionRepository) Subscribe(ctx context.Context, subscription *domain.Subscription) (*domain.Subscription, error) {
	subscription.UnsubscribeToken = uuid.NewString()
	subscription.CreatedAt = time.Now()

	docRef, _, err := sr.db.Collection("subscriptions").Add(ctx, subscription)
	if err != nil {
		return nil, err
	}

	subscription.ID = docRef.ID
	return subscription, nil
}

func (sr *SubscriptionRepository) Unsubscribe(ctx context.Context) error {
	return nil
}
