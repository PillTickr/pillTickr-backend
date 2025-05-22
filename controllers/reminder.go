package controllers

import (
	"context"
	"net/http"
	"pillTickr-backend/db"
	"pillTickr-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetReminders(c *gin.Context) {
	rows, err := db.DB.Query(context.Background(), "SELECT id, user_id, name, notes, is_recurring, recurrence_pattern, start_date, end_date, is_active, created_at FROM pilltickr.reminders")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reminders []models.Reminder

	for rows.Next() {
		var r models.Reminder
		err := rows.Scan(&r.UserID, &r.Name, &r.Notes, &r.IsRecurring, &r.RecurrencePattern, &r.StartDate, &r.EndDate, &r.IsActive, &r.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reminders = append(reminders, r)
	}

	c.JSON(http.StatusOK, reminders)
}

func CreateReminder(c *gin.Context) {
	var reminder models.Reminder
	if err := c.ShouldBindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the reminder into the database
	var insertedID uuid.UUID
	err := db.DB.QueryRow(context.Background(),
		"INSERT INTO pilltickr.reminders (user_id, name, notes, is_recurring, recurrence_pattern, start_date, end_date, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		reminder.UserID,
		reminder.Name,
		reminder.Notes,
		reminder.IsRecurring,
		reminder.RecurrencePattern,
		reminder.StartDate,
		reminder.EndDate,
		reminder.IsActive,
	).Scan(&insertedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, dose := range reminder.Doses {
		_, err := db.DB.Exec(context.Background(),
			"INSERT INTO pilltickr.doses (reminder_id, time, dosage, notes) VALUES ($1, $2, $3, $4)",
			insertedID,
			dose.Time,
			dose.Dosage,
			dose.Notes,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, reminder)
}
