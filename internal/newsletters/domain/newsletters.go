package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Newsletter represents a newsletter object.
type Newsletter struct {
	ID          uuid.UUID `json:"id"`          // ID of the newsletter
	OwnerID     uuid.UUID `json:"owner_id"`    // There is only one owner for each newsletter
	Name        string    `json:"name"`        // Name of the newsletter
	Description string    `json:"description"` // Description of the newsletter
	CreatedAt   time.Time `json:"created_at"`  // Creation time of the newsletter
}

// NewsletterService is an interface that contains a collection of method signatures
// which will be implemented in application level and are responsible for creating a newsletter
// and getting a list of all of them that belong to a particular user.
type NewsletterService interface {
	Create(newsletter *Newsletter) (*Newsletter, error)
	GetAll(ownerID uuid.UUID) ([]*Newsletter, error)
}

// NewsletterRepository is an interface that contains a collection of method signatures
// which will be implemented in persistence level and are responsible for creating a newsletter
// and getting a list of all of them that belong to a particular user.
type NewsletterRepository interface {
	Create(ctx context.Context, newsletter *Newsletter) (*Newsletter, error)
	GetAll(ctx context.Context, ownerID uuid.UUID) ([]*Newsletter, error)
}
