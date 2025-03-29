package utils

import (
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
