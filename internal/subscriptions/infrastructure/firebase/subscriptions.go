package firebase

import (
	"context"
	"errors"
	"newsletter/internal/subscriptions/domain"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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

// Unsubscribe removes a subscription from Firestore based on the unsubscribe token.
//
// It searches the "subscriptions" collection for a document whose "unsubscribeToken"
// field matches the provided token. If a matching document is found, it is deleted.
//
// Parameters:
//   - ctx: Context for controlling cancellation and deadlines for the Firestore operation.
//   - token: The unique unsubscribe token associated with the subscription to be removed.
//
// Returns:
//   - error: Returns an error if no matching subscription is found, or if the Firestore
//     operation fails for any reason.
//
// Notes:
//   - This function only deletes the first subscription found with the given token.
//   - The unsubscribe token should be unique to avoid accidental deletion of multiple subscriptions.
//   - The Firestore field name used in the query is "unsubscribeToken", matching the struct tag in the Subscription entity.
func (sr *SubscriptionRepository) Unsubscribe(ctx context.Context, unsubscribeToken string) error {
	iter := sr.db.
		Collection("subscriptions").
		Where("unsubscribeToken", "==", unsubscribeToken).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return errors.New("subscription not found")
		}
		return err
	}

	_, err = doc.Ref.Delete(ctx)
	return err
}
