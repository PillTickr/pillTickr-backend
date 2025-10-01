package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"pillTickr-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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

type UserResponse struct {
	ID        int       `json:"id"`
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

// ---------- Utils ----------

func generateJWT(userID int, email string) (string, int64, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", 0, fmt.Errorf("JWT_SECRET not set")
	}

	expiration := time.Now().Add(24 * time.Hour).Unix()

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

// ---------- Handlers ----------

// Signup new user
func Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert into users
	var userID int
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

	// Generate JWT
	token, exp, err := generateJWT(userID, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Response
	resp := AuthResponse{
		AccessToken:  token,
		RefreshToken: "", // implement later if needed
		ExpiresIn:    86400,
		TokenType:    "bearer",
		ExpiresAt:    exp,
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

	// Fetch user
	var userID int
	var storedHash, name, email string
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

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT
	token, exp, err := generateJWT(userID, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Response
	resp := AuthResponse{
		AccessToken:  token,
		RefreshToken: "",
		ExpiresIn:    86400,
		TokenType:    "bearer",
		ExpiresAt:    exp,
		User: UserResponse{
			ID:        userID,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
		},
	}
	c.JSON(http.StatusOK, resp)
}
