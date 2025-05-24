package controllers

import (
	"context"
	"net/http"
	"pillTickr-backend/db"
	"pillTickr-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func GetReminders(c *gin.Context) {
	// Extract user ID from JWT claims
	claims, userExists := c.Get("user")
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims, ok := claims.(jwt.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user claims"})
		return
	}

	userID, ok := userClaims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Query reminders for the authenticated user
	rows, err := db.DB.Query(context.Background(),
		"SELECT id, user_id, name, notes, is_recurring, recurrence_pattern, start_date, end_date, is_active, created_at FROM pilltickr.reminders WHERE user_id = $1",
		userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reminders []models.Reminder
	for rows.Next() {
		var r models.Reminder
		err := rows.Scan(&r.ID, &r.UserID, &r.Name, &r.Notes, &r.IsRecurring, &r.RecurrencePattern, &r.StartDate, &r.EndDate, &r.IsActive, &r.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reminders = append(reminders, r)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reminders)
}

func CreateReminder(c *gin.Context) {
	// Extract user ID from JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims, ok := claims.(jwt.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user claims"})
		return
	}

	userID, ok := userClaims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var reminder models.Reminder
	if err := c.ShouldBindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Automatically set user_id from JWT
	reminder.UserID = userID
	// if reminder.id is not set, generate a new UUID
	if reminder.ID == "" {
		reminder.ID = uuid.New().String()
	}
	// reminder.ID = uuid.New().String() // Generate new UUID for reminder

	// Insert the reminder into the database
	var insertedID uuid.UUID
	err := db.DB.QueryRow(context.Background(),
		"INSERT INTO pilltickr.reminders (id, user_id, name, notes, is_recurring, recurrence_pattern, start_date, end_date, is_active, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		reminder.ID,
		reminder.UserID,
		reminder.Name,
		reminder.Notes,
		reminder.IsRecurring,
		reminder.RecurrencePattern,
		reminder.StartDate,
		reminder.EndDate,
		reminder.IsActive,
		reminder.CreatedAt,
	).Scan(&insertedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, dose := range reminder.Doses {
		// Validate time format (HH:MM)
		if _, err := time.Parse("15:04", dose.Time); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format for dose, expected HH:MM"})
			return
		}
		reminder.Doses[i].ReminderID = insertedID.String()
		_, err := db.DB.Exec(context.Background(),
			"INSERT INTO pilltickr.doses (reminder_id, time, dosage, notes, created_at) VALUES ($1, $2, $3, $4, $5)",
			insertedID,
			dose.Time,
			dose.Dosage,
			dose.Notes,
			dose.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, reminder)
}

func UpdateReminder(c *gin.Context) {
	// Extract user ID from JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims, ok := claims.(jwt.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user claims"})
		return
	}

	userID, ok := userClaims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Extract reminder ID from URL
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
		return
	}

	// Bind JSON input to reminder struct
	var reminder models.Reminder
	if err := c.ShouldBindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Automatically set user_id and reminder ID
	reminder.UserID = userID
	reminder.ID = reminderID.String()

	// Verify the reminder exists and belongs to the user
	var reminderExists bool
	err = db.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pilltickr.reminders WHERE id = $1 AND user_id = $2)",
		reminderID, userID).Scan(&reminderExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !reminderExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found or not owned by user"})
		return
	}

	// Update the reminder in the database
	_, err = db.DB.Exec(context.Background(),
		"UPDATE pilltickr.reminders SET name = $1, notes = $2, is_recurring = $3, recurrence_pattern = $4, start_date = $5, end_date = $6, is_active = $7, created_at = $8 WHERE id = $9 AND user_id = $10",
		reminder.Name,
		reminder.Notes,
		reminder.IsRecurring,
		reminder.RecurrencePattern,
		reminder.StartDate,
		reminder.EndDate,
		reminder.IsActive,
		reminder.CreatedAt,
		reminderID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete existing doses for the reminder
	_, err = db.DB.Exec(context.Background(),
		"DELETE FROM pilltickr.doses WHERE reminder_id = $1",
		reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert new doses
	for _, dose := range reminder.Doses {
		// Validate time format (HH:MM)
		if _, err := time.Parse("15:04", dose.Time); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format for dose, expected HH:MM"})
			return
		}
		_, err := db.DB.Exec(context.Background(),
			"INSERT INTO pilltickr.doses (reminder_id, time, dosage, notes, created_at) VALUES ($1, $2, $3, $4, $5)",
			reminderID,
			dose.Time,
			dose.Dosage,
			dose.Notes,
			dose.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, reminder)
}

func DeleteReminder(c *gin.Context) {
	// Extract user ID from JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found"})
		return
	}

	userClaims, ok := claims.(jwt.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user claims"})
		return
	}

	userID, ok := userClaims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// Extract reminder ID from URL
	reminderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
		return
	}

	// Verify the reminder exists and belongs to the user
	var reminderExists bool
	err = db.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pilltickr.reminders WHERE id = $1 AND user_id = $2)",
		reminderID, userID).Scan(&reminderExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !reminderExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found or not owned by user"})
		return
	}

	// Delete associated doses (cascade delete could handle this if set up in DB)
	_, err = db.DB.Exec(context.Background(),
		"DELETE FROM pilltickr.doses WHERE reminder_id = $1",
		reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the reminder
	_, err = db.DB.Exec(context.Background(),
		"DELETE FROM pilltickr.reminders WHERE id = $1 AND user_id = $2",
		reminderID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted reminder with id :" + reminderID.String()})
}
