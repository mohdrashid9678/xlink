package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
	"github.com/mohdrashid9678/xlink/pkg/logger"
)

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, models.Response{Success: true, Data: data})
}

func writeError(c *gin.Context, status int, code, message string) {
	writeProblem(c, status, code, http.StatusText(status), message, nil)
}

func writeProblem(c *gin.Context, status int, code, title, detail string, invalidParams []models.InvalidParam) {
	reqID := logger.GetRequestID(c.Request.Context())
	c.JSON(status, models.ProblemDetails{
		Type:          "urn:xlink:error:" + code,
		Title:         title,
		Status:        status,
		Detail:        detail,
		Code:          code,
		Instance:      c.Request.URL.Path,
		RequestID:     reqID,
		InvalidParams: invalidParams,
	})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeProblem(c, http.StatusBadRequest, "validation_error", "Validation Failed", err.Error(), nil)
	case errors.Is(err, service.ErrNotFound):
		writeProblem(c, http.StatusNotFound, "not_found", "Resource Not Found", "The requested resource was not found.", nil)
	case errors.Is(err, service.ErrConflict):
		writeProblem(c, http.StatusConflict, "conflict", "Resource Conflict", "The requested resource already exists or conflict occurred.", nil)
	default:
		slog.ErrorContext(c.Request.Context(), "Unhandled service error", slog.Any("error", err))
		writeProblem(c, http.StatusInternalServerError, "internal_error", "Internal Server Error", "An unexpected error occurred.", nil)
	}
}
