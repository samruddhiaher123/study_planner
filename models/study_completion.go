package models

import "time"

type StudyCompletion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"-"`
	TaskID      uint      `json:"taskId"`
	CompletedAt time.Time `gorm:"not null;index" json:"completedAt"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
}
