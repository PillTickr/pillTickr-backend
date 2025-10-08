package handlers

import (
	"net/http"
	"time"

	"pillTickr-backend/db"
	"pillTickr-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Input/Output Structs ----------

type SignupInput struct {
	Name     string `json:"name" binding:"required"` // display name
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int          `json:"expiresIn"`
	TokenType    string       `json:"tokenType"`
	ExpiresAt    int64        `json:"expiresAt"`
	User         UserResponse `json:"user"`
}

var accessTokenExpiry = 30 * time.Minute
var refreshTokenExpiry = 7 * 24 * time.Hour

// Register new user
func Register(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	var userID string
	var createdAt time.Time
	err = db.DB.QueryRow(
		`INSERT INTO users (name, email, password_hash, created_at) 
		 VALUES ($1, $2, $3, $4) RETURNING user_id, created_at`,
		input.Name, input.Email, string(hashedPassword), time.Now(),
	).Scan(&userID, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	accessToken, accessExp, err := utils.GenerateJWT(userID, input.Email, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	refreshToken, _, err := utils.GenerateJWT(userID, input.Email, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
		return
	}

	// Optionally store refresh token in DB
	// _, _ = db.DB.Exec(`UPDATE users SET refresh_token = $1 WHERE user_id = $2`, refreshToken, userID)

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenExpiry.Seconds()),
		TokenType:    "bearer",
		ExpiresAt:    accessExp,
		User: UserResponse{
			ID:        userID,
			Name:      input.Name,
			Email:     input.Email,
			CreatedAt: createdAt,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// Login existing user
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var storedHash, name, email, userID string
	var createdAt time.Time
	err := db.DB.QueryRow(
		`SELECT user_id, name, email, password_hash, created_at 
		 FROM users WHERE email = $1`,
		input.Email,
	).Scan(&userID, &name, &email, &storedHash, &createdAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	accessToken, accessExp, err := utils.GenerateJWT(userID, email, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}
	refreshToken, _, err := utils.GenerateJWT(userID, email, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
		return
	}

	// Optionally store refresh token
	_, _ = db.DB.Exec(`UPDATE users SET refresh_token = $1 WHERE user_id = $2`, refreshToken, userID)

	resp := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenExpiry.Seconds()),
		TokenType:    "bearer",
		ExpiresAt:    accessExp,
		User: UserResponse{
			ID:        userID,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
		},
	}
	c.JSON(http.StatusOK, resp)
}

func RefreshToken(c *gin.Context) {
	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, claims, err := utils.ParseJWT(input.RefreshToken)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID := claims["id"].(string)
	email := claims["email"].(string)

	// // Verify token exists in DB (optional but secure)
	// var storedToken string
	// err = db.DB.QueryRow(`SELECT refresh_token FROM users WHERE user_id = $1`, userID).Scan(&storedToken)
	// if err != nil || storedToken != input.RefreshToken {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token mismatch"})
	// 	return
	// }

	// Generate new tokens
	newAccessToken, accessExp, err := utils.GenerateJWT(userID, email, accessTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}
	newRefreshToken, _, err := utils.GenerateJWT(userID, email, refreshTokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
		return
	}

	// Rotate refresh token
	_, _ = db.DB.Exec(`UPDATE users SET refresh_token = $1 WHERE user_id = $2`, newRefreshToken, userID)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
		"expiresIn":    int(accessTokenExpiry.Seconds()),
		"tokenType":    "bearer",
		"expiresAt":    accessExp,
	})
}
