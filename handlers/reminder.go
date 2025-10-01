package handlers

import (
	"database/sql"
	"net/http"
	"pillTickr-backend/db"
	"pillTickr-backend/models"
	"pillTickr-backend/utils"

	"github.com/gin-gonic/gin"
)

func GetReminders(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	rows, err := db.DB.Query(`
		SELECT r.reminder_id, r.schedule_id, r.reminder_datetime, r.status, r.taken_at
		FROM reminders r
		INNER JOIN schedules s ON r.schedule_id = s.schedule_id
		INNER JOIN medicines m ON s.medicine_id = m.medicine_id
		WHERE m.user_id = ?`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reminders []models.Reminder
	for rows.Next() {
		var r models.Reminder
		err := rows.Scan(&r.ID, &r.ScheduleID, &r.ReminderDatetime, &r.Status, &r.TakenAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reminders = append(reminders, r)
	}

	c.JSON(http.StatusOK, reminders)
}

func CreateReminder(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var reminder models.Reminder
	if err := c.ShouldBindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate reminder time
	if reminder.ReminderDatetime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reminder datetime is required"})
		return
	}

	reminder.Status = "pending"

	_, err := db.DB.Exec(`
		INSERT INTO reminders ( schedule_id, reminder_datetime, status, taken_at)
		VALUES ( ?, ?, ?, ?)`,
		reminder.ScheduleID,
		reminder.ReminderDatetime,
		reminder.Status,
		sql.NullTime{Valid: false}, // no taken_at yet
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reminder)
}

func UpdateReminder(c *gin.Context) {
	reminderID := c.Param("id")
	var reminder models.Reminder

	if err := c.ShouldBindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure valid time
	if reminder.ReminderDatetime.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reminder datetime is required"})
		return
	}

	_, err := db.DB.Exec(`
		UPDATE reminders
		SET reminder_datetime = ?, status = ?, taken_at = ?
		WHERE id = ?`,
		reminder.ReminderDatetime,
		reminder.Status,
		reminder.TakenAt,
		reminderID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reminder)
}

func DeleteReminder(c *gin.Context) {
	reminderID := c.Param("id")

	_, err := db.DB.Exec(`DELETE FROM reminders WHERE id = ?`, reminderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
}
