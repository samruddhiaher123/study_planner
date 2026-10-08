package models

import "time"

type Task struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"-"`
	Title       string    `json:"title"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Date        string    `json:"date"`
	StartTime   string    `json:"startTime"`
	EndTime     string    `json:"endTime"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}
