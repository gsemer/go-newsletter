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

// Create inserts a new newsletter record into the database for a user.
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

// GetAll retrieves all newsletters belonging to a specific owner.
func (nr *NewsletterRepository) GetAll(ctx context.Context, ownerID uuid.UUID, limit, page int) ([]*domain.Newsletter, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	query := `select id, owner_id, name, description, created_at from newsletters where owner_id = $1 limit $2 offset $3`

	rows, err := nr.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsletters []*domain.Newsletter
	for rows.Next() {
		var newsletter domain.Newsletter
		err := rows.Scan(
			&newsletter.ID,
			&newsletter.OwnerID,
			&newsletter.Name,
			&newsletter.Description,
			&newsletter.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		newsletters = append(newsletters, &newsletter)
	}

	return newsletters, nil
}
