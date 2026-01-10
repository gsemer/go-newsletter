package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"newsletter/internal/infrastructure/workerpool"
	notifications "newsletter/internal/notifications/domain"
	"newsletter/internal/subscriptions/domain"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock subscription service ---

type MockSubscriptionService struct {
	mock.Mock
}

func (m *MockSubscriptionService) Subscribe(s *domain.Subscription) (*domain.Subscription, error) {
	args := m.Called(s)
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionService) Unsubscribe(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

// -- Mock email service ---

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) Send(email *notifications.Email) error {
	args := m.Called(email)
	return args.Error(0)
}

// Mock job submiter

type MockWorkerPool struct {
	mock.Mock
}

func (m *MockWorkerPool) Submit(job workerpool.Job) {
	m.Called(job)
}

// Tests

func TestSubscribe_Success(t *testing.T) {
	// Arrange
	ss := new(MockSubscriptionService)
	es := new(MockEmailService)
	wp := new(MockWorkerPool)

	h := NewSubscriptionHandler(ss, es, wp)

	sub := &domain.Subscription{
		ID:               "sub-123",
		NewsletterID:     "news-1",
		Email:            "user@test.com",
		UnsubscribeToken: "token-123",
		CreatedAt:        time.Now(),
	}

	ss.On("Subscribe", mock.AnythingOfType("*domain.Subscription")).Return(sub, nil)
	wp.On("Submit", mock.AnythingOfType("*jobs.SendEmailJob")).Return()

	body := map[string]string{"email": "user@test.com"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/news-1", bytes.NewReader(payload))
	req = mux.SetURLVars(req, map[string]string{"newsletter_id": "news-1"})

	rec := httptest.NewRecorder()

	// Act
	h.Subscribe(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp SubscribeResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)

	assert.Equal(t, sub.ID, resp.ID)
	assert.Equal(t, sub.NewsletterID, resp.NewsletterID)
	assert.Equal(t, sub.Email, resp.Email)
	assert.WithinDuration(t, time.Now(), resp.CreatedAt, time.Second)

	ss.AssertExpectations(t)
	wp.AssertExpectations(t)
}
