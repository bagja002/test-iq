package database

import (
	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/auth"
	"test-iq-ku/apps/api/internal/models"
)

func SeedDemoData(db *gorm.DB) error {
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount == 0 {
		adminPassword, err := auth.HashPassword("admin123")
		if err != nil {
			return err
		}

		userPassword, err := auth.HashPassword("user123")
		if err != nil {
			return err
		}

		users := []models.User{
			{
				Name:         "Admin Demo",
				Email:        "admin@testiq.local",
				PasswordHash: adminPassword,
				Role:         models.RoleAdmin,
				Status:       models.UserStatusActive,
			},
			{
				Name:         "Peserta Demo",
				Email:        "user@testiq.local",
				PasswordHash: userPassword,
				Role:         models.RoleUser,
				Status:       models.UserStatusActive,
			},
		}

		if err := db.Create(&users).Error; err != nil {
			return err
		}
	}

	return nil
}
