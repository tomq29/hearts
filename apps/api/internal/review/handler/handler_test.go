package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisssonik/hearts/internal/review"
	"github.com/kisssonik/hearts/internal/review/handler"
	"github.com/kisssonik/hearts/internal/review/service"
	"github.com/kisssonik/hearts/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockReviewService
type MockReviewService struct {
	mock.Mock
}

func (m *MockReviewService) CreateReview(ctx context.Context, authorID string, input service.CreateReviewInput) (*review.Review, error) {
	args := m.Called(ctx, authorID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*review.Review), args.Error(1)
}

func (m *MockReviewService) GetReviewsForUser(ctx context.Context, targetID string) ([]*review.Review, error) {
	args := m.Called(ctx, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*review.Review), args.Error(1)
}

func TestReviewHandler_Create(t *testing.T) {
	mockService := new(MockReviewService)
	logger := zap.NewNop()
	h := handler.NewReviewHandler(mockService, logger)

	input := service.CreateReviewInput{TargetID: "u2", Rating: 5, Comment: "Good"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "u1"))
	w := httptest.NewRecorder()

	expectedReview := &review.Review{ID: "r1", AuthorID: "u1", TargetID: "u2", Rating: 5}
	mockService.On("CreateReview", mock.Anything, "u1", input).Return(expectedReview, nil)

	h.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestReviewHandler_List(t *testing.T) {
	mockService := new(MockReviewService)
	logger := zap.NewNop()
	h := handler.NewReviewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/reviews/u2", nil)
	req.SetPathValue("userID", "u2")
	w := httptest.NewRecorder()

	reviews := []*review.Review{
		{ID: "r1", AuthorID: "u1", TargetID: "u2", Rating: 5},
	}
	mockService.On("GetReviewsForUser", mock.Anything, "u2").Return(reviews, nil)

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*review.Review
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
}
