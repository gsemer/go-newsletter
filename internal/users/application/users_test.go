package application

import (
	"context"
	"errors"
	"newsletter/internal/users/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ------------------- Mocks -------------------

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Get(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// ------------------- Tests -------------------

func TestUserService_Create_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	us := NewUserService(mockRepo)

	inputUser := &domain.User{Email: "test@example.com", Password: "hashed"}
	createdUser := &domain.User{ID: uuid.New(), Email: "test@example.com"}

	mockRepo.On("Create", mock.Anything, inputUser).Return(createdUser, nil)

	result, err := us.Create(inputUser)

	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Create_Failure(t *testing.T) {
	mockRepo := new(MockUserRepository)
	us := NewUserService(mockRepo)

	inputUser := &domain.User{Email: "fail@example.com", Password: "hashed"}

	mockRepo.On("Create", mock.Anything, inputUser).Return((*domain.User)(nil), errors.New("create failed"))

	result, err := us.Create(inputUser)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// ------------------- Authenticate -------------------

func TestAuthenticationService_Authenticate_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	as := NewAuthenticationService(mockRepo)

	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	storedUser := &domain.User{ID: uuid.New(), Email: "test@example.com", Password: string(hashed)}

	mockRepo.On("Get", mock.Anything, "test@example.com").Return(storedUser, nil)

	user, err := as.Authenticate("test@example.com", password)

	assert.NoError(t, err)
	assert.Equal(t, storedUser.ID, user.ID)
	assert.Equal(t, "", user.Password, "password should be cleared")
	mockRepo.AssertExpectations(t)
}

func TestAuthenticationService_Authenticate_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	as := NewAuthenticationService(mockRepo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	storedUser := &domain.User{ID: uuid.New(), Email: "test@example.com", Password: string(hashed)}

	mockRepo.On("Get", mock.Anything, "test@example.com").Return(storedUser, nil)

	user, err := as.Authenticate("test@example.com", "wrongpass")

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticationService_Authenticate_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	as := NewAuthenticationService(mockRepo)

	mockRepo.On("Get", mock.Anything, "missing@example.com").Return((*domain.User)(nil), errors.New("not found"))

	user, err := as.Authenticate("missing@example.com", "any")

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

// ------------------- GenerateAccessToken -------------------

func TestAuthenticationService_GenerateAccessToken_Success(t *testing.T) {
	as := &AuthenticationService{}
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	// Set a temporary JWT_SECRET_KEY for test
	t.Setenv("JWT_SECRET_KEY", "secret123")

	token, err := as.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthenticationService_GenerateAccessToken_Failure(t *testing.T) {
	as := &AuthenticationService{}
	user := &domain.User{
		ID:    uuid.Nil, // invalid ID still works, but we'll test secret missing
		Email: "test@example.com",
	}

	// Unset JWT_SECRET_KEY to simulate signing failure
	t.Setenv("JWT_SECRET_KEY", "")

	token, err := as.GenerateAccessToken(user)

	assert.Error(t, err)
	assert.Equal(t, "", token)
}
