package application_test

import (
	"context"
	"errors"
	"newsletter/internal/subscriptions/application"
	"newsletter/internal/subscriptions/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repository ---
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Subscribe(ctx context.Context, s *domain.Subscription) (*domain.Subscription, error) {
	args := m.Called(ctx, s)
	sub := args.Get(0)
	if sub == nil {
		return nil, args.Error(1)
	}
	return sub.(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Unsubscribe(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// --- Tests for Subscribe ---

func TestSubscribe_Success(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	subscription := &domain.Subscription{
		NewsletterID: "newsletter1",
		Email:        "test@example.com",
	}

	createdSub := &domain.Subscription{
		ID:           "sub123",
		NewsletterID: subscription.NewsletterID,
		Email:        subscription.Email,
	}

	// Expect repository Subscribe to be called
	mockRepo.On("Subscribe", mock.Anything, subscription).Return(createdSub, nil)

	result, err := ss.Subscribe(subscription)

	assert.NoError(t, err)
	assert.Equal(t, createdSub, result)

	mockRepo.AssertExpectations(t)
}

func TestSubscribe_Failure(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	subscription := &domain.Subscription{
		NewsletterID: "newsletter1",
		Email:        "fail@example.com",
	}

	mockRepo.On("Subscribe", mock.Anything, subscription).Return(nil, errors.New("db error"))

	result, err := ss.Subscribe(subscription)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.EqualError(t, err, "db error")

	mockRepo.AssertExpectations(t)
}

// --- Tests for Unsubscribe ---

func TestUnsubscribe_Success(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	token := "token123"

	mockRepo.On("Unsubscribe", mock.Anything, token).Return(nil)

	err := ss.Unsubscribe(token)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUnsubscribe_Failure(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	token := "token123"

	mockRepo.On("Unsubscribe", mock.Anything, token).Return(errors.New("not found"))

	err := ss.Unsubscribe(token)

	assert.Error(t, err)
	assert.EqualError(t, err, "not found")
	mockRepo.AssertExpectations(t)
}

// --- Timeout / context test (optional, ensures context is used) ---
func TestSubscribe_ContextTimeout(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	subscription := &domain.Subscription{
		NewsletterID: "newsletter1",
		Email:        "timeout@example.com",
	}

	// Simulate long-running operation
	mockRepo.On("Subscribe", mock.Anything, subscription).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done() // block until context is cancelled
	}).Return(nil, context.DeadlineExceeded)

	start := time.Now()
	_, err := ss.Subscribe(subscription)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(6000)) // context timeout ~5s
	mockRepo.AssertExpectations(t)
}

func TestUnsubscribe_ContextTimeout(t *testing.T) {
	mockRepo := new(MockSubscriptionRepository)
	ss := application.NewSubscriptionService(mockRepo)

	token := "timeouttoken"

	mockRepo.On("Unsubscribe", mock.Anything, token).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done() // block until context is cancelled
	}).Return(context.DeadlineExceeded)

	start := time.Now()
	err := ss.Unsubscribe(token)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(6000))
	mockRepo.AssertExpectations(t)
}
