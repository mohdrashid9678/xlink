package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
)

type AnalyticsHandler struct {
	service service.AnalyticsService
}

func NewAnalyticsHandler(service service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "A valid access token is required.", nil)
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		writeProblem(c, http.StatusBadRequest, "invalid_request", "Invalid Request", "Short code is required.", nil)
		return
	}

	query := models.AnalyticsQuery{
		Interval: c.DefaultQuery("interval", "day"),
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			query.From = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			query.To = t
		}
	}

	summary, err := h.service.GetURLAnalytics(c.Request.Context(), userID, shortCode, query)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, summary)
}
