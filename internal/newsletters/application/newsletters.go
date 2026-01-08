package application

import (
	"context"
	"log/slog"
	"newsletter/internal/newsletters/domain"
	"time"

	"github.com/google/uuid"
)

// NewsletterService provides application-level operations related to newsletters
// and it orchestrates domain logic and persistence concerns.
type NewsletterService struct {
	nr domain.NewsletterRepository
}

func NewNewsletterService(nr domain.NewsletterRepository) *NewsletterService {
	return &NewsletterService{nr: nr}
}

// Create creates a new newsletter.
//
// This method applies application-level orchestration, including logging
// and execution time limits. It persists the provided newsletter through
// the repository and returns the newly created newsletter populated with
// persistence-related fields (such as ID and creation timestamp).
//
// A context with a fixed timeout is used to prevent the operation from
// blocking indefinitely.
func (ns *NewsletterService) Create(newsletter *domain.Newsletter) (*domain.Newsletter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	slog.Info(
		"creating newsletter",
		"owner_id", newsletter.OwnerID,
		"name", newsletter.Name,
	)

	newNewsletter, err := ns.nr.Create(ctx, newsletter)
	if err != nil {
		slog.Error(
			"failed to create newsletter",
			"owner_id", newsletter.OwnerID,
			"name", newsletter.Name,
			"error", err,
		)
		return nil, err
	}

	return newNewsletter, nil
}

func (ns *NewsletterService) GetAll(ownerID uuid.UUID) ([]*domain.Newsletter, error) {
	return []*domain.Newsletter{}, nil
}
