package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"pillTickr-backend/db"

	"github.com/gin-gonic/gin"
)

type SignupInput struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	DOB         string `json:"dob" binding:"required"` // Format: YYYY-MM-DD, will be parsed to time.Time
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Age         int       `json:"age"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// calculateAge computes age from DOB (time.Time) based on today's date
func calculateAge(dob time.Time) int {
	today := time.Now()
	age := today.Year() - dob.Year()
	if today.Month() < dob.Month() || (today.Month() == dob.Month() && today.Day() < dob.Day()) {
		age--
	}
	return age
}

func Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse DOB string to time.Time
	dobTime, err := time.Parse("2006-01-02", input.DOB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DOB format, expected YYYY-MM-DD"})
		return
	}

	// Calculate age
	age := calculateAge(dobTime)

	// Supabase signup
	authURL := os.Getenv("SUPABASE_URL") + "/auth/v1/signup"
	apiKey := os.Getenv("SUPABASE_ANON_KEY")

	signupPayload := map[string]interface{}{
		"email":    input.Email,
		"password": input.Password,
	}
	signupBody, err := json.Marshal(signupPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal signup payload"})
		return
	}

	signupReq, err := http.NewRequest("POST", authURL, bytes.NewBuffer(signupBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create signup request"})
		return
	}
	signupReq.Header.Set("apikey", apiKey)
	signupReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	signupResp, err := client.Do(signupReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send signup request"})
		return
	}
	defer signupResp.Body.Close()

	if signupResp.StatusCode != http.StatusOK && signupResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(signupResp.Body)
		c.JSON(signupResp.StatusCode, gin.H{"error": string(respBody)})
		return
	}

	var signupResult struct {
		User struct {
			ID        string    `json:"id"`
			Email     string    `json:"email"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"user"`
	}
	if err := json.NewDecoder(signupResp.Body).Decode(&signupResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse signup response"})
		return
	}

	// Insert into profiles table
	_, err = db.DB.Exec(context.Background(),
		`INSERT INTO pilltickr.profiles (id, display_name, dob, created_at)
		VALUES ($1, $2, $3, $4)`,
		signupResult.User.ID, input.DisplayName, dobTime, signupResult.User.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert profile", "details": err.Error()})
		return
	}

	// Login to get access token
	loginURL := os.Getenv("SUPABASE_URL") + "/auth/v1/token?grant_type=password"
	loginPayload := map[string]interface{}{
		"email":    input.Email,
		"password": input.Password,
	}
	loginBody, err := json.Marshal(loginPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal login payload"})
		return
	}

	loginReq, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create login request"})
		return
	}
	loginReq.Header.Set("apikey", apiKey)
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login request failed"})
		return
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(loginResp.Body)
		c.JSON(loginResp.StatusCode, gin.H{"error": string(respBody)})
		return
	}

	var loginResult AuthResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&loginResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse login response"})
		return
	}

	// Modify user response
	loginResult.User = UserResponse{
		ID:          signupResult.User.ID,
		Email:       signupResult.User.Email,
		DisplayName: input.DisplayName,
		Age:         age,
		CreatedAt:   signupResult.User.CreatedAt,
	}

	respBody, err := json.Marshal(loginResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Supabase login
	authURL := os.Getenv("SUPABASE_URL") + "/auth/v1/token?grant_type=password"
	apiKey := os.Getenv("SUPABASE_ANON_KEY")

	payload := map[string]interface{}{
		"email":    input.Email,
		"password": input.Password,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal login payload"})
		return
	}

	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create login request"})
		return
	}
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to Supabase"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(respBody)})
		return
	}

	var loginResult AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse login response"})
		return
	}

	// Fetch profile data
	var profile struct {
		DisplayName string    `json:"display_name"`
		DOB         time.Time `json:"dob"`
		CreatedAt   time.Time `json:"created_at"`
	}
	err = db.DB.QueryRow(context.Background(),
		`SELECT display_name, dob, created_at FROM pilltickr.profiles WHERE id = $1`,
		loginResult.User.ID).Scan(&profile.DisplayName, &profile.DOB, &profile.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile", "details": err.Error()})
		return
	}

	// Calculate age
	age := calculateAge(profile.DOB)

	// Modify user response
	loginResult.User = UserResponse{
		ID:          loginResult.User.ID,
		Email:       loginResult.User.Email,
		DisplayName: profile.DisplayName,
		Age:         age,
		CreatedAt:   profile.CreatedAt,
	}

	respBody, err := json.Marshal(loginResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", respBody)
}
