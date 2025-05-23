package routes

import (
	"pillTickr-backend/controllers"
	"pillTickr-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/auth/signup", controllers.Signup)
		api.POST("/auth/login", controllers.Login)

		auth := api.Group("/")
		auth.Use(middleware.RequireAuth())
		auth.GET("/reminders", controllers.GetReminders)
		auth.POST("/reminders", controllers.CreateReminder)
		auth.PUT("/reminders/:id", controllers.UpdateReminder)
		auth.DELETE("/reminders/:id", controllers.DeleteReminder)
	}
}
