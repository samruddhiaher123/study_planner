package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"study-planner/dto"
	"study-planner/models"
)

//search controller which accept the query parameter

func parseTaskFilter(r *http.Request) (models.TaskFilter, error) {
	params := r.URL.Query()
	filter := dto.TaskListQuery{
		Search:   strings.TrimSpace(params.Get("q")),
		Subject:  strings.TrimSpace(params.Get("subject")),
		Status:   params.Get("status"),
		Priority: params.Get("priority"),
		Date:     params.Get("date"),
		From:     params.Get("from"),
		To:       params.Get("to"),
		View:     params.Get("view"),
		Sort:     params.Get("sort"),
	}
	if filter.Status == "Pending" {
		filter.Status = string(models.TaskNotStarted)
	}
	if filter.Status != "" && !models.TaskStatus(filter.Status).IsValid() {
		return models.TaskFilter{}, fmt.Errorf("Invalid status filter")
	}
	if filter.Priority != "" && !models.TaskPriority(filter.Priority).IsValid() {
		return models.TaskFilter{}, fmt.Errorf("Invalid priority filter")
	}
	//for date must be into date-m-y format
	for _, key := range []string{"date", "from", "to"} {
		value := params.Get(key)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return models.TaskFilter{}, fmt.Errorf("%s must use YYYY-MM-DD", key)
		}
	}
	if filter.From != "" && filter.To != "" && filter.From > filter.To {
		return models.TaskFilter{}, fmt.Errorf("from must be on or before to")
	}
	switch filter.View {
	case "", "all", "completed", "pending", "overdue", "upcoming":
	default:
		return models.TaskFilter{}, fmt.Errorf("Invalid task view")
	}
	switch filter.Sort {
	case "", "default", "dueDate", "priority", "newest", "oldest", "status":
	default:
		return models.TaskFilter{}, fmt.Errorf("Invalid sort option")
	}
	return filter.ToTaskFilter(), nil
}

func validateTask(input *dto.CreateTaskRequest) string {
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
		input.Status = string(models.TaskNotStarted)
	}
	if !models.TaskStatus(input.Status).IsValid() {
		return "Status must be Not Started, In Progress, Completed, or On Hold"
	}
	if input.Priority == nil || *input.Priority == "" {
		defaultPriority := string(models.PriorityMedium)
		input.Priority = &defaultPriority
	}
	if !models.TaskPriority(*input.Priority).IsValid() {
		return "Priority must be High, Medium, or Low"
	}
	return ""
}
