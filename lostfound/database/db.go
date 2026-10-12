package database

import (
	"log"
	"os"

	"lostfound/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=" + os.Getenv("USER") +
			" dbname=lostfound port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Could not connect to database: ", err)
	}

	// Creates the tables automatically
	err = db.AutoMigrate(&models.User{}, &models.Item{})
	if err != nil {
		log.Fatal("Could not create tables: ", err)
	}

	DB = db
	log.Println("Database connected")
}