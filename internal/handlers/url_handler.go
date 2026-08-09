package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	var request models.CreateURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	url, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, url)
}

func (h *URLHandler) Get(c *gin.Context) {
	shortCode := c.Param("shortCode")
	url, err := h.service.Get(c.Request.Context(), shortCode)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, url)
}

// Redirect sends visitors from a public short code to its destination. The
// management API's Get handler intentionally remains a non-counted read.
func (h *URLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	url, err := h.service.Get(c.Request.Context(), shortCode)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.Redirect(http.StatusFound, url.LongURL)
	h.trackClickAsync(shortCode)
}

func (h *URLHandler) Update(c *gin.Context) {
	var request models.UpdateURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}

	url, err := h.service.Update(c.Request.Context(), c.Param("shortCode"), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, url)
}

func (h *URLHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("shortCode")); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"deleted": true})
}
