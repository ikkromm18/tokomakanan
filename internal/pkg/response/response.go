package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ikkromm18/tokomakanan/internal/dto"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBusinessRule = errors.New("business rule violation")
)

type ServiceError struct {
	Type    error
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(errType error, message string) error {
	return &ServiceError{Type: errType, Message: message}
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, dto.StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func Paginated(c *gin.Context, message string, data any, meta dto.PaginationMeta) {
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, dto.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func ValidationError(c *gin.Context, err error) {
	var valErrors []dto.FieldError
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			valErrors = append(valErrors, dto.FieldError{
				Field:   fe.Field(),
				Message: fe.Tag() + " validation failed",
			})
		}
	}
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  valErrors,
	})
}

func HandleServiceError(c *gin.Context, err error) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		switch {
		case errors.Is(serviceErr.Type, ErrNotFound):
			Error(c, http.StatusNotFound, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrDuplicate):
			Error(c, http.StatusConflict, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrUnauthorized):
			Error(c, http.StatusUnauthorized, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrForbidden):
			Error(c, http.StatusForbidden, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrBusinessRule):
			Error(c, http.StatusUnprocessableEntity, serviceErr.Message)
		default:
			Error(c, http.StatusBadRequest, serviceErr.Message)
		}
		return
	}
	Error(c, http.StatusInternalServerError, "Internal server error")
}
