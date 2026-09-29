package main

import (
	"fmt"
	"log"
	"os"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/database"
	"suppliment/backend/internal/models"
	"suppliment/backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("unable to connect database: %v", err)
	}
	name := os.Getenv("ADMIN_NAME")
	if name == "" {
		name = "System Administrator"
	}
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@suppliment.local"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "Admin@12345"
	}
	ensureAdmin(db, name, email, password)
	fmt.Println("Admin account ready.")
}

func ensureAdmin(db *gorm.DB, name, email, password string) {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err == nil {
		log.Printf("Admin already exists: %s", user.Email)
		return
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash error: %v", err)
	}
	user = models.User{ID: uuid.New(), Name: name, Email: email, PasswordHash: hash, Role: models.RoleAdmin}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("create admin failed: %v", err)
	}
	fmt.Printf("Created admin %s (%s)\n", name, email)
}
