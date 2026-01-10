package application

import (
	"context"
	"newsletter/internal/subscriptions/domain"
	"time"
)

type SubscriptionService struct {
	sr domain.SubscriptionRepository
}

func NewSubscriptionService(sr domain.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{sr: sr}
}

// Subscribe creates a new subscription for a given newsletter.
//
// Parameters:
//   - subscription: pointer to a Subscription domain object containing
//     the newsletter ID and subscriber email.
//
// Returns:
//   - pointer to the created Subscription object (with ID, timestamps, etc. populated)
//   - error if the subscription could not be created
//
// Behavior:
//   - Uses a context with a 5-second timeout to ensure the operation does not hang.
//   - Delegates the actual persistence to the subscription repository.
func (ss *SubscriptionService) Subscribe(subscription *domain.Subscription) (*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newSubscription, err := ss.sr.Subscribe(ctx, subscription)
	if err != nil {
		return nil, err
	}

	return newSubscription, nil
}

func (ss *SubscriptionService) Unsubscribe() error {
	return nil
}
