package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/internal/service"
)

type AuthHandler struct{ service service.AuthService }

func NewAuthHandler(service service.AuthService) *AuthHandler { return &AuthHandler{service: service} }

func (h *AuthHandler) Register(c *gin.Context) {
	var request models.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	response, err := h.service.Register(c.Request.Context(), request)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeSuccess(c, http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, response)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var request models.RefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	response, err := h.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var request models.LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON.")
		return
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		writeAuthError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"logged_out": true})
}

func writeAuthError(c *gin.Context, err error) {
	if err == service.ErrInvalidCredentials {
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials or refresh token.")
		return
	}
	writeServiceError(c, err)
}
