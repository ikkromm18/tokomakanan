package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
)

type sampleStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Success(c, http.StatusOK, "Operation succeeded", map[string]string{"foo": "bar"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Operation succeeded", resp.Message)
	dataMap, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "bar", dataMap["foo"])
}

func TestCreatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Created(c, "Resource created", map[string]int{"id": 10})

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Resource created", resp.Message)
}

func TestNoContentResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.NoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.Bytes())
}

func TestPaginatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	meta := dto.PaginationMeta{
		Page:       1,
		Limit:      10,
		TotalRows:  25,
		TotalPages: 3,
	}
	items := []string{"item1", "item2"}

	response.Paginated(c, "Items retrieved", items, meta)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Items retrieved", resp.Message)
	assert.Equal(t, meta.Page, resp.Meta.Page)
	assert.Equal(t, meta.Limit, resp.Meta.Limit)
	assert.Equal(t, meta.TotalRows, resp.Meta.TotalRows)
	assert.Equal(t, meta.TotalPages, resp.Meta.TotalPages)
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusBadRequest, "Invalid parameter")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Invalid parameter", resp.Message)
	assert.Empty(t, resp.Errors)
}

func TestValidationError(t *testing.T) {
	validate := validator.New()
	sample := sampleStruct{Name: "", Email: "invalid-email"}
	err := validate.Struct(sample)
	assert.Error(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.ValidationError(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	unmarshalErr := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, unmarshalErr)
	assert.False(t, resp.Success)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.NotEmpty(t, resp.Errors)
}

func TestHandleServiceError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "ErrNotFound maps to 404",
			err:          response.NewServiceError(response.ErrNotFound, "Product not found"),
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Product not found",
		},
		{
			name:         "ErrDuplicate maps to 409",
			err:          response.NewServiceError(response.ErrDuplicate, "Email already exists"),
			expectedCode: http.StatusConflict,
			expectedMsg:  "Email already exists",
		},
		{
			name:         "ErrUnauthorized maps to 401",
			err:          response.NewServiceError(response.ErrUnauthorized, "Invalid token"),
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "Invalid token",
		},
		{
			name:         "ErrForbidden maps to 403",
			err:          response.NewServiceError(response.ErrForbidden, "Access denied"),
			expectedCode: http.StatusForbidden,
			expectedMsg:  "Access denied",
		},
		{
			name:         "ErrBusinessRule maps to 422",
			err:          response.NewServiceError(response.ErrBusinessRule, "Cannot cancel paid order"),
			expectedCode: http.StatusUnprocessableEntity,
			expectedMsg:  "Cannot cancel paid order",
		},
		{
			name:         "Unknown ServiceError type maps to 400",
			err:          response.NewServiceError(errors.New("other_error"), "Bad request error"),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Bad request error",
		},
		{
			name:         "Generic error maps to 500",
			err:          errors.New("unexpected database crash"),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			response.HandleServiceError(c, tt.err)

			assert.Equal(t, tt.expectedCode, w.Code)
			var resp dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.False(t, resp.Success)
			assert.Equal(t, tt.expectedMsg, resp.Message)
		})
	}
}

func TestServiceError_ErrorMethod(t *testing.T) {
	err := response.NewServiceError(response.ErrNotFound, "Not found message")
	assert.Equal(t, "Not found message", err.Error())
}
