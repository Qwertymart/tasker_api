package model

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	CreatedAt   *time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsReady     bool       `json:"is ready"`
}

func ConnectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=5432 dbname=task_manager port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		return nil, err
	}

	return db, nil
}
