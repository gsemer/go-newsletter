package application

import (
	"context"
	"log/slog"
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

	slog.Info("Creating subscription", "newsletter_id", subscription.NewsletterID, "email", subscription.Email)

	newSubscription, err := ss.sr.Subscribe(ctx, subscription)
	if err != nil {
		slog.Error(
			"Failed to create subscription",
			"newsletter_id", subscription.NewsletterID,
			"email", subscription.Email,
			"error", err,
		)
		return nil, err
	}

	slog.Info(
		"Subscription created successfully",
		"subscription_id", newSubscription.ID,
		"newsletter_id", newSubscription.NewsletterID,
		"email", newSubscription.Email,
	)

	return newSubscription, nil
}

// Unsubscribe removes a subscription associated with the given unsubscribe token.
//
// This method is part of the SubscriptionService and acts as the application-level
// logic for handling unsubscription requests. It delegates the deletion to the
// underlying repository while enforcing a timeout.
//
// Parameters:
//   - unsubscribeToken: A unique token identifying the subscription to remove.
//
// Behavior:
//   - Creates a context with a 5-second timeout for the repository operation.
//   - Calls the SubscriptionRepository's Unsubscribe method to delete the subscription.
//   - Returns any error encountered during the deletion, or nil if successful.
func (ss *SubscriptionService) Unsubscribe(unsubscribeToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Attempting to unsubscribe", "token", unsubscribeToken)

	err := ss.sr.Unsubscribe(ctx, unsubscribeToken)
	if err != nil {
		slog.Error("Failed to unsubscribe", "token", unsubscribeToken, "error", err)
		return err
	}

	slog.Info("Unsubscribed successfully", "token", unsubscribeToken)
	return nil
}
