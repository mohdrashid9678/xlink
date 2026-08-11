package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohdrashid9678/xlink/internal/auth"
	"github.com/mohdrashid9678/xlink/internal/models"
)

const UserIDContextKey = "authenticated_user_id"

func RequireAuth(jwt *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}
		userID, err := jwt.Verify(parts[1])
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(UserIDContextKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(UserIDContextKey)
	userID, ok := value.(uuid.UUID)
	return userID, exists && ok
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{Success: false, Error: &models.ErrorInfo{Code: "unauthorized", Message: "A valid access token is required."}})
}
