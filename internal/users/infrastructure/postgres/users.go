package postgres

import (
	"context"
	"database/sql"
	"newsletter/internal/users/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository implements persistence operations for domain.User entities
// using a PostgreSQL database.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create persists a new user in the database.
//
// The user's plaintext password is hashed using bcrypt before storage.
// On success, Create returns a fully initialized domain.User containing
// the generated ID, email, and creation timestamp.
//
// The returned user will never contain a password or password hash.
//
// Possible errors include:
//   - bcrypt hashing failures
//   - database constraint violations (e.g. duplicate email)
//   - database connectivity errors
func (ur *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return nil, err
	}

	var userDB *domain.User = &domain.User{}
	query := `insert into users (password, email, created_at) values ($1, $2, $3) returning id, email, created_at`

	err = ur.db.QueryRowContext(
		ctx,
		query,
		hashedPassword,
		user.Email,
		time.Now(),
	).Scan(&userDB.ID, &userDB.Email, &userDB.CreatedAt)
	if err != nil {
		return nil, err
	}

	userDB.Password = ""

	return userDB, nil
}

// Get retrieves a user by email address.
//
// The returned user includes the stored password hash, making this method
// suitable for authentication-related use cases.
//
// If no user exists with the given email, Get returns an error (typically sql.ErrNoRows).
func (ur *UserRepository) Get(ctx context.Context, email string) (*domain.User, error) {
	query := `select id, password, email, created_at from users where email = $1`

	var user *domain.User = &domain.User{}
	err := ur.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Password, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}
