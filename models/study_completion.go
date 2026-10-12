package models

import "time"

type StudyCompletion struct {
	ID                 uint      `gorm:"primaryKey"`
	UserID             uint      `gorm:"not null;index"`
	TaskID             uint      `gorm:"not null;index"`
	ActualStudyMinutes int       `gorm:"not null;default:0"`
	CompletedAt        time.Time `gorm:"not null;index"`
}
