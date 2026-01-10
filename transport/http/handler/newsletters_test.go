package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"newsletter/internal/newsletters/domain"
	userdomain "newsletter/internal/users/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Newsletter Service ---
type MockNewsletterService struct {
	mock.Mock
}

func (m *MockNewsletterService) Create(n *domain.Newsletter) (*domain.Newsletter, error) {
	args := m.Called(n)
	return args.Get(0).(*domain.Newsletter), args.Error(1)
}

func (m *MockNewsletterService) GetAll(ownerID uuid.UUID, limit, page int) ([]*domain.Newsletter, error) {
	args := m.Called(ownerID, limit, page)
	return args.Get(0).([]*domain.Newsletter), args.Error(1)
}

// --- helper function to set user ID in context ---
func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userdomain.UserID, userID)
}

// --- Tests ---

func TestCreateNewsletter_Success(t *testing.T) {
	mockSvc := new(MockNewsletterService)
	h := NewNewsletterHandler(mockSvc)

	ownerID := uuid.New()
	body := domain.Newsletter{Name: "Tech Newsletter"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/newsletters", bytes.NewReader(jsonBody))
	req = req.WithContext(contextWithUserID(req.Context(), ownerID.String()))
	rec := httptest.NewRecorder()

	created := &domain.Newsletter{ID: uuid.New(), OwnerID: ownerID, Name: body.Name}
	mockSvc.On("Create", mock.AnythingOfType("*domain.Newsletter")).Return(created, nil)

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp domain.Newsletter
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, resp.ID)

	mockSvc.AssertExpectations(t)
}

func TestCreateNewsletter_Unauthorized(t *testing.T) {
	mockSvc := new(MockNewsletterService)
	h := NewNewsletterHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/newsletters", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetAllNewsletters_Success(t *testing.T) {
	mockSvc := new(MockNewsletterService)
	h := NewNewsletterHandler(mockSvc)

	ownerID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/newsletters?limit=2&page=1", nil)
	req = req.WithContext(contextWithUserID(req.Context(), ownerID.String()))
	rec := httptest.NewRecorder()

	newsletters := []*domain.Newsletter{
		{ID: uuid.New(), OwnerID: ownerID, Name: "Tech"},
		{ID: uuid.New(), OwnerID: ownerID, Name: "Science"},
	}

	mockSvc.On("GetAll", ownerID, 2, 1).Return(newsletters, nil)

	h.GetAll(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []*domain.Newsletter
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)

	mockSvc.AssertExpectations(t)
}
