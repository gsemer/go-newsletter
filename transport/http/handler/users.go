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
// Route:
//
//	POST /users/signup
//
// Description:
//
//	Registers a new user using an email and password. If registration
//	succeeds, an access token is generated and returned in the
//	"Authorization" response header.
//
// Request Body (application/json):
//
//	{
//	  "email": "user@example.com",
//	  "password": "password"
//	}
//
// Responses:
//
//	201 Created
//	  Headers:
//	    Authorization: Bearer <access_token>
//	  Body:
//	    {
//	      "id": "uuid",
//	      "email": "user@example.com",
//	      "created_at": "2026-01-10T12:00:00Z"
//	    }
//
//	400 Bad Request
//	  - Invalid JSON payload
//	  - User creation failure (e.g. validation errors)
//
//	500 Internal Server Error
//	  - Token generation failure
//
// Side Effects:
//   - Persists a new user record
//   - Generates an access token for authentication
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

// Signin handles user authentication.
//
// Route:
//
//	POST /users/signin
//
// Description:
//
//	Authenticates a user using email and password. On success, an access
//	token is returned in the "Authorization" response header and the
//	authenticated user is returned in the response body.
//
// Request Body (application/json):
//
//	{
//	  "email": "user@example.com",
//	  "password": "password"
//	}
//
// Responses:
//
//	200 OK
//	  Headers:
//	    Authorization: Bearer <access_token>
//	  Body:
//	    {
//	      "id": "uuid",
//	      "email": "user@example.com",
//	      "created_at": "2026-01-10T12:00:00Z"
//	    }
//
//	400 Bad Request
//	  - Invalid JSON payload
//
//	401 Unauthorized
//	  - Invalid email or password
//
//	500 Internal Server Error
//	  - Token generation failure
//
// Side Effects:
//   - Generates a new access token
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
