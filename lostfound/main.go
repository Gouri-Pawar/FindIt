package main

import (
	"os"

	"lostfound/auth"
	"lostfound/database"
	"lostfound/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	os.MkdirAll("uploads", 0755)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MB

	// The website and the uploaded photos
	r.StaticFile("/", "./web/index.html")
	r.Static("/uploads", "./uploads")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Open routes: anyone can use
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.GET("/items", handlers.ListItems)
	r.GET("/items/:id", handlers.GetItem)

	// Protected routes: token needed
	protected := r.Group("/")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/me", handlers.Me)
		protected.POST("/items", handlers.CreateItem)
		protected.PUT("/items/:id", handlers.UpdateItem)
		protected.DELETE("/items/:id", handlers.DeleteItem)
	}

	r.Run(":8080")
}