package main

import (
	"pillTickr-backend/db"
	"pillTickr-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	r := gin.Default()
	routes.RegisterRoutes(r)

	r.Run(":8080")
}
