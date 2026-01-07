package application

import (
	"context"
	"log/slog"
	"newsletter/config"
	"newsletter/internal/users/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserService provides application-level operations related to users
// and it orchestrates domain logic and persistence concerns.
type UserService struct {
	ur domain.UserRepository
}

func NewUserService(ur domain.UserRepository) *UserService {
	return &UserService{ur: ur}
}

// Create registers a new user in the system.
//
// A timeout is applied to the operation to prevent long-running database
// calls from blocking the request lifecycle.
//
// On success, Create returns the newly created user entity.
// On failure, the error is logged and returned to the caller.
func (us *UserService) Create(user *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	slog.Info(
		"creating user",
		"email", user.Email,
	)

	newUser, err := us.ur.Create(ctx, user)
	if err != nil {
		slog.Error(
			"failed to create user",
			"email", user.Email,
			"error", err,
		)
		return nil, err
	}

	return newUser, nil
}

type AuthenticationService struct {
	ur domain.UserRepository
}

func NewAuthenticationService(ur domain.UserRepository) *AuthenticationService {
	return &AuthenticationService{ur: ur}
}

// Authenticate verifies a user's credentials by email and password.
//
// It returns the authenticated user if credentials are valid.
// The user's password hash is cleared before returning to prevent accidental exposure.
func (us *UserService) Authenticate(email, password string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, err := us.ur.Get(ctx, email)
	if err != nil {
		slog.Error("failed to find user",
			"email", email,
			"error", err,
		)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		slog.Warn("invalid password attempt",
			"email", email,
		)
		return nil, err
	}

	// Never expose password hash
	user.Password = ""

	slog.Info("user authenticated successfully",
		"user_id", user.ID.String(),
		"email", user.Email,
	)

	return user, nil
}

// GenerateAccessToken generates a JWT access token for an authenticated user.
// The token is short-lived (15 minutes) and includes the user's email and ID.
func (us *UserService) GenerateAccessToken(user *domain.User) (string, error) {
	slog.Info("generating access token",
		"user_id", user.ID.String(),
		"email", user.Email,
	)

	claims := &domain.Claims{
		Email: user.Email,
		RegisteredClaims: &jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := access.SignedString([]byte(config.GetEnv("JWT_SECRET_KEY", "")))
	if err != nil {
		slog.Error("failed to sign access token",
			"user_id", user.ID.String(),
			"error", err,
		)
		return "", err
	}

	slog.Info("access token generated successfully",
		"user_id", user.ID.String(),
	)

	return accessToken, nil
}
