package domain

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Custom type for context keys to avoid collisions
type ContextKey string

const (
	UserID ContextKey = "userID"
)

// User represents the user account.
type User struct {
	ID        uuid.UUID `json:"id,omitempty"` // ID of the user
	Password  string    `json:"password"`     // Hashed password of the user
	Email     string    `json:"email"`        // Email of the user
	CreatedAt time.Time `json:"created_at"`   // Creation time of the user
}

// UserService is an interface that contains a collection of method signatures
// which will be implemented in application level and are responsible for creating a user.
type UserService interface {
	Create(user *User) (*User, error)
}

// UserRepository is an interface that contains a collection of method signatures
// which will be implemented in persistence level and are responsible for creating
// and getting a user.
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Get(ctx context.Context, email string) (*User, error)
}

type Claims struct {
	Email string
	*jwt.RegisteredClaims
}

// AuthService is an interface that contains a collection of method signatures
// which will be implemented in application level and are responsible for authenticating a user
// and generating a token on sign up/sign in.
type AuthenticationService interface {
	Authenticate(email, password string) (*User, error)
	GenerateAccessToken(user *User) (string, error)
}
