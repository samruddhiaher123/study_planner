package models

import "time"

type TaskStatus string

const (
	TaskNotStarted TaskStatus = "Not Started"
	TaskInProgress TaskStatus = "In Progress"
	TaskCompleted  TaskStatus = "Completed"
	TaskOnHold     TaskStatus = "On Hold"
)

func (status TaskStatus) IsValid() bool {
	switch status {
	case TaskNotStarted, TaskInProgress, TaskCompleted, TaskOnHold:
		return true
	default:
		return false
	}
}

type TaskPriority string

const (
	PriorityHigh   TaskPriority = "High"
	PriorityMedium TaskPriority = "Medium"
	PriorityLow    TaskPriority = "Low"
)

func (priority TaskPriority) IsValid() bool {
	switch priority {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	UserID      uint         `json:"-"`
	Title       string       `json:"title"`
	Subject     string       `json:"subject"`
	Description string       `json:"description"`
	Date        string       `json:"date"`
	StartTime   string       `json:"startTime"`
	EndTime     string       `json:"endTime"`
	Status      TaskStatus   `gorm:"not null;default:Not Started" json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority    TaskPriority `gorm:"not null;default:Medium" json:"priority" enums:"High,Medium,Low"`
	CompletedAt *time.Time   `json:"completedAt,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	User        User         `gorm:"foreignKey:UserID" json:"-"`
}
