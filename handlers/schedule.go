// handlers/schedule.go
package handlers

import (
	"net/http"
	"pillTickr-backend/db"

	"github.com/gin-gonic/gin"
)

// GET /medicines/:id/schedules
func GetSchedules(c *gin.Context) {
	medicineID := c.Param("id")

	rows, err := db.DB.Query(`SELECT schedule_id, start_date, end_date, frequency, times_per_day
		FROM schedules WHERE medicine_id = ?`, medicineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
		return
	}
	defer rows.Close()

	schedules := []gin.H{}
	for rows.Next() {
		var s struct {
			ID          int     `json:"schedule_id"`
			StartDate   string  `json:"start_date"`
			EndDate     *string `json:"end_date"`
			Frequency   string  `json:"frequency"`
			TimesPerDay int     `json:"times_per_day"`
		}
		if err := rows.Scan(&s.ID, &s.StartDate, &s.EndDate, &s.Frequency, &s.TimesPerDay); err == nil {
			schedules = append(schedules, gin.H{
				"id": s.ID, "start_date": s.StartDate, "end_date": s.EndDate,
				"frequency": s.Frequency, "times_per_day": s.TimesPerDay,
			})
		}
	}

	c.JSON(http.StatusOK, schedules)
}

// POST /medicines/:id/schedules
func CreateSchedule(c *gin.Context) {
	medicineID := c.Param("id")

	var req struct {
		StartDate   string  `json:"start_date" binding:"required"`
		EndDate     *string `json:"end_date"`
		Frequency   string  `json:"frequency" binding:"required"`
		TimesPerDay int     `json:"times_per_day" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := db.DB.Exec(`INSERT INTO schedules (medicine_id, start_date, end_date, frequency, times_per_day)
		VALUES (?, ?, ?, ?, ?)`, medicineID, req.StartDate, req.EndDate, req.Frequency, req.TimesPerDay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"schedule_id": id})
}
