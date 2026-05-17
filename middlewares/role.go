package middlewares

import (
	"github.com/gin-gonic/gin"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesInterface, exists := c.Get("roles")
		if !exists {
			c.JSON(403, gin.H{"error": "Roles tidak ditemukan di context"})
			c.Abort()
			return
		}

		roles, ok := rolesInterface.([]interface{})
		if !ok {
			c.JSON(403, gin.H{"error": "Format roles tidak valid"})
			c.Abort()
			return
		}

		userHasRole := false
		for _, role := range roles {
			roleStr, ok := role.(string)
			if !ok {
				continue
			}

			for _, allowed := range allowedRoles {
				if roleStr == allowed {
					userHasRole = true
					break
				}
			}

			if userHasRole {
				break
			}
		}

		if !userHasRole {
			c.JSON(403, gin.H{"error": "Anda tidak memiliki akses ke resource ini"})
			c.Abort()
			return
		}

		c.Next()
	}
}
