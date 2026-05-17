package helpers

import (
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, roles []string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"roles":   roles,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(config.JwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
