package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
)

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, models.Response{Success: true, Data: data})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, models.Response{
		Success: false,
		Error:   &models.ErrorInfo{Code: code, Message: message},
	})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, service.ErrNotFound):
		writeError(c, http.StatusNotFound, "not_found", "The short URL was not found.")
	case errors.Is(err, service.ErrConflict):
		writeError(c, http.StatusConflict, "short_code_taken", "The requested short code is already in use.")
	default:
		log.Printf("URL request failed: %v", err)
		writeError(c, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
	}
}
