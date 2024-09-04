package handlers

import (
	"time"
	"github.com/dgrijalva/jwt-go"
)

// Secret keys for different roles
var superUserSecretKey = []byte("superuser_secret_key")
var adminSecretKey = []byte("admin_secret_key")
var providerSecretKey = []byte("provider_secret_key")
var userSecretKey = []byte("user_secret_key")

// Token oluşturma fonksiyonları
func GenerateSuperUserToken(userID string) (string, error) {
	return generateToken(userID, "superuser", superUserSecretKey)
}

func GenerateAdminToken(userID string) (string, error) {
	return generateToken(userID, "admin", adminSecretKey)
}

func GenerateProviderToken(userID string) (string, error) {
	return generateToken(userID, "provider", providerSecretKey)
}

func GenerateUserToken(userID string) (string, error) {
	return generateToken(userID, "user", userSecretKey)
}

func generateToken(userID, role string, secretKey []byte) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"exp":  expirationTime.Unix(),
		"sub":  userID,
		"role": role, // Token'a rol bilgisini ekler
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
