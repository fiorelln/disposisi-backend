package middlewares

import (
	"strings"

	"github.com/fiorelln/disposisi/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header tidak ditemukan"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Format Authorization header salah"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return config.JwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Token tidak valid atau expired"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{"error": "Tidak dapat membaca claims dari token"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "user_id tidak ditemukan di token"})
			c.Abort()
			return
		}

		roles, ok := claims["roles"].([]interface{})
		if !ok {
			c.JSON(401, gin.H{"error": "roles tidak ditemukan di token"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))
		c.Set("roles", roles)

		c.Next()
	}
}
