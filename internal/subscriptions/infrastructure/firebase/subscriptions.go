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
