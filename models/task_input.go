package models

type TaskInput struct {
	Title       string
	Subject     string
	Description string
	Date        string
	StartTime   string
	EndTime     string
	Status      TaskStatus
	Priority    *TaskPriority
}

type TaskFilter struct {
	Search   string
	Subject  string
	Status   string
	Priority string
	Date     string
	From     string
	To       string
	View     string
	Sort     string
}
