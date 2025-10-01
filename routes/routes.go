package routes

import (
	"net/http"
	"pillTickr-backend/handlers"
	"pillTickr-backend/middleware"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
	Secured     bool
}

type Routes []Route

func NewRoutes() Routes {
	return Routes{
		// --- Auth ---
		{
			Name:        "Signup",
			Method:      "POST",
			Pattern:     "/auth/signup",
			HandlerFunc: handlers.Signup,
			Secured:     false,
		},
		{
			Name:        "Login",
			Method:      "POST",
			Pattern:     "/auth/login",
			HandlerFunc: handlers.Login,
			Secured:     false,
		},

		// --- Reminders (secured) ---
		{
			Name:        "GetReminders",
			Method:      "GET",
			Pattern:     "/reminders",
			HandlerFunc: handlers.GetReminders,
			Secured:     true,
		},
		{
			Name:        "CreateReminder",
			Method:      "POST",
			Pattern:     "/reminders",
			HandlerFunc: handlers.CreateReminder,
			Secured:     true,
		},
		{
			Name:        "PatchReminder",
			Method:      "PATCH",
			Pattern:     "/reminders/:id",
			HandlerFunc: handlers.UpdateReminder,
			Secured:     true,
		},
		{
			Name:        "DeleteReminder",
			Method:      "DELETE",
			Pattern:     "/reminders/:id",
			HandlerFunc: handlers.DeleteReminder,
			Secured:     true,
		},
		{
			Name:        "GetMedicines",
			Method:      "GET",
			Pattern:     "/medicines",
			HandlerFunc: handlers.GetMedicines,
			Secured:     true,
		},
		{
			Name:        "CreateMedicine",
			Method:      "POST",
			Pattern:     "/medicines",
			HandlerFunc: handlers.CreateMedicine,
			Secured:     true,
		},
		{
			Name:        "GetSchedules",
			Method:      "GET",
			Pattern:     "/medicines/:id/schedules",
			HandlerFunc: handlers.GetSchedules,
			Secured:     true,
		},
		{
			Name:        "CreateSchedule",
			Method:      "POST",
			Pattern:     "/medicines/:id/schedules",
			HandlerFunc: handlers.CreateSchedule,
			Secured:     true,
		},
		{
			Name:        "AddScheduleTime",
			Method:      "POST",
			Pattern:     "/schedules/:id/times",
			HandlerFunc: handlers.AddScheduleTime,
			Secured:     true,
		},
		// --- Health Check ---
		{
			Name:    "HealthCheck",
			Method:  "GET",
			Pattern: "/health",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
			Secured: false,
		},
	}
}

// Attach routes to the server
func AttachRoutes(server *gin.RouterGroup, routes Routes) {
	for _, route := range routes {
		if route.Secured {
			// Wrap with RequireAuth middleware
			server.Handle(route.Method, route.Pattern, middleware.RequireAuth(), route.HandlerFunc)
		} else {
			server.Handle(route.Method, route.Pattern, route.HandlerFunc)
		}
	}
}
