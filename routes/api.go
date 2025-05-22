package routes

import (
	"pillTickr-backend/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/reminders", controllers.GetReminders)
		api.POST("/reminders", controllers.CreateReminder)
		// More routes: PUT, DELETE
	}
}
