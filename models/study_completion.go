package models

import "time"

type StudyCompletion struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	UserID             uint      `gorm:"not null;index" json:"-"`
	TaskID             uint      `json:"taskId"`
	ActualStudyMinutes int       `gorm:"not null;default:0" json:"actualStudyMinutes"`
	CompletedAt        time.Time `gorm:"not null;index" json:"completedAt"`
	User               User      `gorm:"foreignKey:UserID" json:"-"`
}
