package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/models"
	"github.com/mohdrashid9678/xlink/pkg/logger"
)

const (
	UserIDContextKey = "user_id"
	authHeaderPrefix = "Bearer "
)

func RequireAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Authorization header is required.")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, authHeaderPrefix)
		if tokenString == authHeader || tokenString == "" {
			abortWithProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Authorization header must be formatted as 'Bearer <token>'.")
			return
		}

		userID, err := jwtManager.Verify(tokenString)
		if err != nil {
			abortWithProblem(c, http.StatusUnauthorized, "unauthorized", "Unauthorized", "Access token is invalid or expired.")
			return
		}

		c.Set(UserIDContextKey, userID)
		c.Next()
	}
}

func abortWithProblem(c *gin.Context, status int, code, title, detail string) {
	reqID := logger.GetRequestID(c.Request.Context())
	c.AbortWithStatusJSON(status, models.ProblemDetails{
		Type:      "urn:xlink:error:" + code,
		Title:     title,
		Status:    status,
		Detail:    detail,
		Code:      code,
		Instance:  c.Request.URL.Path,
		RequestID: reqID,
	})
}

func UserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(UserIDContextKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}
