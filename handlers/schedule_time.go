// handlers/schedule_time.go
package handlers

import (
	"net/http"
	"pillTickr-backend/db"

	"github.com/gin-gonic/gin"
)

// POST /schedules/:id/times
func AddScheduleTime(c *gin.Context) {
	scheduleID := c.Param("id")

	var req struct {
		IntakeTime string `json:"intake_time" binding:"required"` // "HH:MM"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := db.DB.Exec(`INSERT INTO schedule_times (schedule_id, intake_time) VALUES (?, ?)`, scheduleID, req.IntakeTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add schedule time"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"time_id": id})
}
