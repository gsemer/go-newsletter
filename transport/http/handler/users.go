package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"newsletter/internal/users/domain"
	"time"

	"github.com/google/uuid"
)

type UserHandler struct {
	us domain.UserService
	as domain.AuthenticationService
}

func NewUserHandler(us domain.UserService, as domain.AuthenticationService) *UserHandler {
	return &UserHandler{us: us, as: as}
}

type SignupRequest struct {
	Password string `json:"password"` // Password of the user
	Email    string `json:"email"`    // Email of the user
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// SignUp handles user registration.
//
// It expects a JSON payload with user details (email, password, etc.).
// On success, it returns the created user (without password) in the response body
// and sets the access token in the "Authorization" header in the form "Bearer <token>".
// This allows clients to use the token for subsequent authenticated requests.
func (uh *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var request SignupRequest

	// Decode incoming JSON
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("failed to decode request body", "error", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	user := domain.User{
		Password: request.Password,
		Email:    request.Email,
	}

	// Create user via application service
	newUser, err := uh.us.Create(&user)
	if err != nil {
		slog.Error("failed to create user", "email", user.Email, "error", err)
		http.Error(w, "failed to create user", http.StatusBadRequest)
		return
	}

	newUser.Password = ""

	// Generate access token
	accessToken, err := uh.as.GenerateAccessToken(newUser)
	if err != nil {
		slog.Error("failed to generate access token", "user_id", newUser.ID.String(), "error", err)
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Set token in Authorization header
	w.Header().Set("Authorization", "Bearer "+accessToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := UserResponse{
		ID:        newUser.ID,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
	}

	// Return the created user in response body
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "user_id", newUser.ID.String(), "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("user signed up successfully",
		"user_id", newUser.ID.String(),
		"email", newUser.Email,
	)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signin handles user login (authentication).
//
// It expects a JSON payload with the user's email and password.
// On successful authentication, it returns the authenticated user (without password)
// in the response body and sets the access token in the "Authorization" header
// in the form "Bearer <token>". This token can then be used for subsequent authenticated requests.
func (uh *UserHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	// Decode incoming JSON
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("failed to decode login request", "error", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	slog.Debug("login attempt", "email", request.Email)

	// Authenticate user via application service
	authUser, err := uh.as.Authenticate(request.Email, request.Password)
	if err != nil {
		slog.Warn("authentication failed", "email", request.Email, "error", err)
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	authUser.Password = ""

	slog.Info("user authenticated successfully", "user_id", authUser.ID.String(), "email", authUser.Email)

	// Generate access token
	accessToken, err := uh.as.GenerateAccessToken(authUser)
	if err != nil {
		slog.Error("failed to generate access token", "user_id", authUser.ID.String(), "error", err)
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Set token in Authorization header
	w.Header().Set("Authorization", "Bearer "+accessToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := UserResponse{
		ID:        authUser.ID,
		Email:     authUser.Email,
		CreatedAt: authUser.CreatedAt,
	}

	// Return the authenticated user in response body
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode login response", "user_id", authUser.ID.String(), "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
