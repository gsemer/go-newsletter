package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"newsletter/internal/users/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService mocks domain.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	return args.Get(0).(*domain.User), args.Error(1)
}

// MockAuthService mocks domain.AuthenticationService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Authenticate(email, password string) (*domain.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GenerateAccessToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

// ------------------- SignUp Tests -------------------

func TestUserHandler_SignUp_Success(t *testing.T) {
	mockUS := new(MockUserService)
	mockAS := new(MockAuthService)

	handler := &UserHandler{
		us: mockUS,
		as: mockAS,
	}

	inputUser := &domain.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	createdUser := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	mockUS.On("Create", inputUser).Return(createdUser, nil)
	mockAS.On("GenerateAccessToken", createdUser).Return("token123", nil)

	body, _ := json.Marshal(inputUser)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.SignUp(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "Bearer token123", resp.Header.Get("Authorization"))

	var respUser domain.User
	json.NewDecoder(resp.Body).Decode(&respUser)
	assert.Equal(t, createdUser.ID, respUser.ID)

	mockUS.AssertExpectations(t)
	mockAS.AssertExpectations(t)
}

func TestUserHandler_SignUp_CreateUserError(t *testing.T) {
	mockUS := new(MockUserService)
	mockAS := new(MockAuthService)

	handler := &UserHandler{
		us: mockUS,
		as: mockAS,
	}

	inputUser := &domain.User{
		Email:    "fail@example.com",
		Password: "password123",
	}

	mockUS.On("Create", inputUser).Return((*domain.User)(nil), errors.New("create failed"))

	body, _ := json.Marshal(inputUser)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.SignUp(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockUS.AssertExpectations(t)
}

// ------------------- Signin Tests -------------------

func TestUserHandler_Signin_Success(t *testing.T) {
	mockUS := new(MockUserService)
	mockAS := new(MockAuthService)

	handler := &UserHandler{
		us: mockUS,
		as: mockAS,
	}

	input := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	authUser := &domain.User{
		ID:    uuid.New(),
		Email: input.Email,
	}

	mockAS.On("Authenticate", input.Email, input.Password).Return(authUser, nil)
	mockAS.On("GenerateAccessToken", authUser).Return("token123", nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer token123", resp.Header.Get("Authorization"))

	var respUser domain.User
	json.NewDecoder(resp.Body).Decode(&respUser)
	assert.Equal(t, authUser.ID, respUser.ID)

	mockAS.AssertExpectations(t)
}

func TestUserHandler_Signin_AuthFailed(t *testing.T) {
	mockUS := new(MockUserService)
	mockAS := new(MockAuthService)

	handler := &UserHandler{
		us: mockUS,
		as: mockAS,
	}

	input := LoginRequest{
		Email:    "fail@example.com",
		Password: "wrongpass",
	}

	mockAS.On("Authenticate", input.Email, input.Password).Return((*domain.User)(nil), errors.New("auth failed"))

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	mockAS.AssertExpectations(t)
}
