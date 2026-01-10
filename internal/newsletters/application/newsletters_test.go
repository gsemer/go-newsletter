package application_test

import (
	"context"
	"errors"
	"newsletter/internal/newsletters/application"
	"newsletter/internal/newsletters/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Newsletter Repository ---
type MockNewsletterRepository struct {
	mock.Mock
}

func (m *MockNewsletterRepository) Create(ctx context.Context, n *domain.Newsletter) (*domain.Newsletter, error) {
	args := m.Called(ctx, n)
	news := args.Get(0)
	if news == nil {
		return nil, args.Error(1)
	}
	return news.(*domain.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) GetAll(ctx context.Context, ownerID uuid.UUID, limit, page int) ([]*domain.Newsletter, error) {
	args := m.Called(ctx, ownerID, limit, page)
	news := args.Get(0)
	if news == nil {
		return nil, args.Error(1)
	}
	return news.([]*domain.Newsletter), args.Error(1)
}

// --- Tests for Create ---

func TestCreateNewsletter_Success(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	newsletter := &domain.Newsletter{
		OwnerID: uuid.New(),
		Name:    "Tech News",
	}

	created := &domain.Newsletter{
		ID:      uuid.New(),
		OwnerID: newsletter.OwnerID,
		Name:    newsletter.Name,
	}

	mockRepo.On("Create", mock.Anything, newsletter).Return(created, nil)

	result, err := ns.Create(newsletter)

	assert.NoError(t, err)
	assert.Equal(t, created, result)

	mockRepo.AssertExpectations(t)
}

func TestCreateNewsletter_Failure(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	newsletter := &domain.Newsletter{
		OwnerID: uuid.New(),
		Name:    "Fail Newsletter",
	}

	mockRepo.On("Create", mock.Anything, newsletter).Return(nil, errors.New("db error"))

	result, err := ns.Create(newsletter)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.EqualError(t, err, "db error")

	mockRepo.AssertExpectations(t)
}

// Timeout / context test
func TestCreateNewsletter_ContextTimeout(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	newsletter := &domain.Newsletter{
		OwnerID: uuid.New(),
		Name:    "Timeout Newsletter",
	}

	mockRepo.On("Create", mock.Anything, newsletter).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(nil, context.DeadlineExceeded)

	start := time.Now()
	_, err := ns.Create(newsletter)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(2000)) // 1s timeout + small overhead
	mockRepo.AssertExpectations(t)
}

// --- Tests for GetAll ---

func TestGetAllNewsletters_Success(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	ownerID := uuid.New()
	newsletters := []*domain.Newsletter{
		{ID: uuid.New(), OwnerID: ownerID, Name: "Tech"},
		{ID: uuid.New(), OwnerID: ownerID, Name: "Science"},
	}

	mockRepo.On("GetAll", mock.Anything, ownerID, 10, 1).Return(newsletters, nil)

	result, err := ns.GetAll(ownerID, 10, 1)

	assert.NoError(t, err)
	assert.Equal(t, newsletters, result)

	mockRepo.AssertExpectations(t)
}

func TestGetAllNewsletters_Failure(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	ownerID := uuid.New()

	mockRepo.On("GetAll", mock.Anything, ownerID, 10, 1).Return(nil, errors.New("db error"))

	result, err := ns.GetAll(ownerID, 10, 1)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.EqualError(t, err, "db error")

	mockRepo.AssertExpectations(t)
}

// Timeout / context test
func TestGetAllNewsletters_ContextTimeout(t *testing.T) {
	mockRepo := new(MockNewsletterRepository)
	ns := application.NewNewsletterService(mockRepo)

	ownerID := uuid.New()

	mockRepo.On("GetAll", mock.Anything, ownerID, 10, 1).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(nil, context.DeadlineExceeded)

	start := time.Now()
	_, err := ns.GetAll(ownerID, 10, 1)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.LessOrEqual(t, elapsed.Milliseconds(), int64(1000)) // 500ms + small overhead
	mockRepo.AssertExpectations(t)
}
