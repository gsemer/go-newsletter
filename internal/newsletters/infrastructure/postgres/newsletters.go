package postgres

import (
	"context"
	"database/sql"
	"newsletter/internal/newsletters/domain"
	"time"

	"github.com/google/uuid"
)

type NewsletterRepository struct {
	db *sql.DB
}

func NewNewsletterRepository(db *sql.DB) *NewsletterRepository {
	return &NewsletterRepository{db: db}
}

func (nr *NewsletterRepository) Create(ctx context.Context, newsletter *domain.Newsletter) (*domain.Newsletter, error) {
	var newsletterDB *domain.Newsletter = &domain.Newsletter{}
	query := `insert into newsletters (owner_id, name, description, created_at) values ($1, $2, $3, $4) returning id, owner_id, name, description, created_at`

	err := nr.db.QueryRowContext(
		ctx,
		query,
		newsletter.OwnerID,
		newsletter.Name,
		newsletter.Description,
		time.Now(),
	).Scan(&newsletterDB.ID, &newsletterDB.OwnerID, &newsletterDB.Name, &newsletterDB.Description, &newsletterDB.CreatedAt)
	if err != nil {
		return nil, err
	}

	return newsletterDB, nil
}

func (nr *NewsletterRepository) GetAll(ctx context.Context, ownerID uuid.UUID) ([]*domain.Newsletter, error) {
	return []*domain.Newsletter{}, nil
}
