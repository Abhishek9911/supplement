package middleware

import (
	"net/http"
	"strings"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/models"
	"suppliment/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthRequired(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(token, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
			return
		}
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
			return
		}
		c.Set("user_id", userID)
		c.Set("user_role", models.Role(claims.Role))
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}
		role, ok := roleValue.(models.Role)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid role context"})
			return
		}
		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
	}
}

func CurrentUserID(c *gin.Context) uuid.UUID {
	value, exists := c.Get("user_id")
	if !exists { return uuid.Nil }
	userID, ok := value.(uuid.UUID)
	if !ok { return uuid.Nil }
	return userID
}

func CurrentRole(c *gin.Context) models.Role {
	value, exists := c.Get("user_role")
	if !exists { return models.RoleCustomer }
	role, ok := value.(models.Role)
	if !ok { return models.RoleCustomer }
	return role
}
