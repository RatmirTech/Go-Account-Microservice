package middleware

import (
	"net/http"
	"strings"

	"AccountService/internal/token"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := tm.ParseAccess(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
