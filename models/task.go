package models

import "time"

type Task struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	UserID             uint       `json:"-"`
	Title              string     `json:"title"`
	Subject            string     `json:"subject"`
	Description        string     `json:"description"`
	Date               string     `json:"date"`
	StartTime          string     `json:"startTime"`
	EndTime            string     `json:"endTime"`
	Status             string     `gorm:"not null;default:Not Started" json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority           string     `gorm:"not null;default:Medium" json:"priority" enums:"High,Medium,Low"`
	ActualStudyMinutes int        `gorm:"not null;default:0" json:"actualStudyMinutes"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}
