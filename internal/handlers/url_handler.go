package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/middleware"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
)

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (h *URLHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "A valid access token is required.", nil)
		return
	}
	var request models.CreateURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_request", "Invalid Request", "Request body must be valid JSON.", nil)
		return
	}

	url, err := h.service.Create(c.Request.Context(), userID, request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, url)
}

func (h *URLHandler) Get(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "A valid access token is required.", nil)
		return
	}
	shortCode := c.Param("shortCode")
	url, err := h.service.Get(c.Request.Context(), userID, shortCode)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, url)
}

func (h *URLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	url, err := h.service.GetPublic(c.Request.Context(), shortCode)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.Redirect(http.StatusFound, url.LongURL)
	// h.trackClickAsync(shortCode)
}

func (h *URLHandler) Update(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "A valid access token is required.", nil)
		return
	}
	var request models.UpdateURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeProblem(c, http.StatusBadRequest, "invalid_request", "Invalid Request", "Request body must be valid JSON.", nil)
		return
	}

	url, err := h.service.Update(c.Request.Context(), userID, c.Param("shortCode"), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, url)
}

func (h *URLHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		writeProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "A valid access token is required.", nil)
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, c.Param("shortCode")); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"deleted": true})
}
