package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"study-planner/middleware"
)

type taskRequest struct {
	Title              string  `json:"title"`
	Subject            string  `json:"subject"`
	Description        string  `json:"description"`
	Date               string  `json:"date"`
	StartTime          string  `json:"startTime"`
	EndTime            string  `json:"endTime"`
	Status             string  `json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority           *string `json:"priority" enums:"High,Medium,Low"`
	ActualStudyMinutes *int    `json:"actualStudyMinutes" minimum:"0" maximum:"1440"`
}

type messageResponse struct {
	Message string `json:"message"`
}

const taskSelect = `
	SELECT id, user_id, title, subject, description, date, start_time, end_time,
	       status, priority, actual_study_minutes, completed_at, created_at
	FROM tasks
`

func filteredTaskQuery(r *http.Request) (string, []any, error) {
	conditions := []string{"user_id = ?"}
	args := []any{middleware.UserID(r)}
	params := r.URL.Query()

	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, condition)
	}

	if search := strings.TrimSpace(params.Get("q")); search != "" {
		args = append(args, "%"+search+"%", "%"+search+"%")
		conditions = append(conditions, "(title LIKE ? OR description LIKE ?)")
	}
	if subject := strings.TrimSpace(params.Get("subject")); subject != "" {
		addCondition("subject = ?", subject)
	}
	if status := params.Get("status"); status != "" {
		if status == "Pending" {
			status = "Not Started"
		}
		if !validTaskStatus(status) {
			return "", nil, fmt.Errorf("Invalid status filter")
		}
		addCondition("status = ?", status)
	}
	if priority := params.Get("priority"); priority != "" {
		if priority != "High" && priority != "Medium" && priority != "Low" {
			return "", nil, fmt.Errorf("Invalid priority filter")
		}
		addCondition("priority = ?", priority)
	}

	for _, key := range []string{"date", "from", "to"} {
		value := params.Get(key)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return "", nil, fmt.Errorf("%s must use YYYY-MM-DD", key)
		}
		switch key {
		case "date":
			args = append(args, value)
			conditions = append(conditions, "date = ?")
		case "from":
			args = append(args, value)
			conditions = append(conditions, "date >= ?")
		case "to":
			args = append(args, value)
			conditions = append(conditions, "date <= ?")
		}
	}
	if from, to := params.Get("from"), params.Get("to"); from != "" && to != "" && from > to {
		return "", nil, fmt.Errorf("from must be on or before to")
	}

	switch params.Get("view") {
	case "", "all":
	case "completed":
		conditions = append(conditions, "status = 'Completed'")
	case "pending":
		conditions = append(conditions, "status <> 'Completed'")
	case "overdue":
		args = append(args, time.Now().Format("2006-01-02"))
		conditions = append(conditions, "date < ? AND status <> 'Completed'")
	case "upcoming":
		args = append(args, time.Now().Format("2006-01-02"))
		conditions = append(conditions, "date >= ? AND status <> 'Completed'")
	default:
		return "", nil, fmt.Errorf("Invalid task view")
	}

	query := taskSelect + " WHERE " + strings.Join(conditions, " AND ")
	today := time.Now().Format("2006-01-02")
	switch params.Get("sort") {
	case "", "default", "dueDate":
		query += " ORDER BY CASE WHEN status = 'Completed' THEN 1 ELSE 0 END,"
		query += " CASE WHEN status <> 'Completed' AND date < '" + today + "' THEN 0 ELSE 1 END,"
		query += " CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END,"
		query += " date, start_time, id"
	case "priority":
		query += " ORDER BY CASE WHEN status = 'Completed' THEN 1 ELSE 0 END,"
		query += " CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END,"
		query += " date, start_time, id"
	case "newest":
		query += " ORDER BY created_at DESC, id DESC"
	case "oldest":
		query += " ORDER BY created_at ASC, id ASC"
	case "status":
		query += " ORDER BY status, date, start_time, id"
	default:
		return "", nil, fmt.Errorf("Invalid sort option")
	}
	return query, args, nil
}

func validTaskStatus(status string) bool {
	return status == "Not Started" || status == "In Progress" || status == "Completed" || status == "On Hold"
}

func validateTask(input *taskRequest) string {
	input.Title = strings.TrimSpace(input.Title)
	input.Subject = strings.TrimSpace(input.Subject)
	if input.Title == "" || input.Subject == "" {
		return "Title and subject are required"
	}
	if len([]rune(input.Title)) > 255 || len([]rune(input.Subject)) > 150 {
		return "Title or subject is too long"
	}
	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return "Date must use YYYY-MM-DD"
	}
	start, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		return "Start time must use HH:MM"
	}
	end, err := time.Parse("15:04", input.EndTime)
	if err != nil || !end.After(start) {
		return "End time must use HH:MM and be later than start time"
	}
	if input.Status == "" || input.Status == "Pending" {
		input.Status = "Not Started"
	}
	if !validTaskStatus(input.Status) {
		return "Status must be Not Started, In Progress, Completed, or On Hold"
	}
	if input.Priority == nil || *input.Priority == "" {
		defaultPriority := "Medium"
		input.Priority = &defaultPriority
	}
	if *input.Priority != "High" && *input.Priority != "Medium" && *input.Priority != "Low" {
		return "Priority must be High, Medium, or Low"
	}
	if input.ActualStudyMinutes != nil && (*input.ActualStudyMinutes < 0 || *input.ActualStudyMinutes > 1440) {
		return "Actual study minutes must be between 0 and 1440"
	}
	return ""
}
