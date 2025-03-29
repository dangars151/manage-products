package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"manage-products/models"
	"os"
	"time"
)

func GenerateToken(user models.User) (string, error) {
	JWT_SECRET := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT_SECRET))
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	JWT_SECRET := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
