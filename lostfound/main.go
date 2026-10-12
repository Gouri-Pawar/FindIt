package main

import (
	"lostfound/auth"
	"lostfound/database"
	"lostfound/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Open routes, anyone can use
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected routes, token needed
	protected := r.Group("/")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/me", handlers.Me)
	}

	r.Run(":8080")
}