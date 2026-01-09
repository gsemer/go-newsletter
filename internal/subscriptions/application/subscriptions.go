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

func (ss *SubscriptionService) Subscribe(subscription *domain.Subscription) (*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
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
