package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func GetUserID(c *gin.Context) (float64, bool) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return 0, false
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user claims"})
		return 0, false
	}

	userID, ok := mapClaims["id"]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return 0, false
	}
	return userID.(float64), true
}

func GenerateJWT(userID string, email string, duration time.Duration) (string, int64, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", 0, fmt.Errorf("JWT_SECRET not set")
	}

	expiration := time.Now().Add(duration).Unix()
	claims := jwt.MapClaims{
		"id":    userID,
		"email": email,
		"exp":   expiration,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signedToken, expiration, nil
}

func ParseJWT(tokenStr string) (*jwt.Token, jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid token claims")
	}
	return token, claims, nil
}
