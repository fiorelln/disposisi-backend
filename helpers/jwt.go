package helpers

import (
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"roles":   roles,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(config.JwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}