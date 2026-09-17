package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPublicService struct {
	mock.Mock
}

func (m *mockPublicService) GetInvoiceByToken(ctx context.Context, token string) (*dto.PublicInvoiceResponse, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublicInvoiceResponse), args.Error(1)
}

func setupPublicRouter(h *handler.PublicHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1/public")
	api.GET("/invoice/:token", h.GetInvoiceByToken)

	return r
}

func TestPublicHandler_GetInvoiceByToken_Success(t *testing.T) {
	svc := new(mockPublicService)
	h := handler.NewPublicHandler(svc)
	router := setupPublicRouter(h)

	token := "4c13a2d0-9999-4e78-bc45-989f66e01a8f"
	expectedResp := &dto.PublicInvoiceResponse{
		InvoiceNo:   "INV/20260917/00001",
		TotalAmount: 50000,
		CreatedAt:   time.Now(),
	}

	svc.On("GetInvoiceByToken", mock.Anything, token).Return(expectedResp, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/public/invoice/"+token, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestPublicHandler_GetInvoiceByToken_NotFound(t *testing.T) {
	svc := new(mockPublicService)
	h := handler.NewPublicHandler(svc)
	router := setupPublicRouter(h)

	svc.On("GetInvoiceByToken", mock.Anything, "missing-token").
		Return(nil, response.NewServiceError(response.ErrNotFound, "invoice not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/public/invoice/missing-token", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}
