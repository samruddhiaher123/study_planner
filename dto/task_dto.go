package dto

import (
	"time"

	"study-planner/models"
)

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Subject     string  `json:"subject"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	Status      string  `json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority    *string `json:"priority" enums:"High,Medium,Low"`
}

type UpdateTaskRequest CreateTaskRequest

type TaskListQuery struct {
	Search   string `query:"q"`
	Subject  string `query:"subject"`
	Status   string `query:"status"`
	Priority string `query:"priority"`
	Date     string `query:"date"`
	From     string `query:"from"`
	To       string `query:"to"`
	View     string `query:"view"`
	Sort     string `query:"sort"`
}

type TaskResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Subject     string     `json:"subject"`
	Description string     `json:"description"`
	Date        string     `json:"date"`
	StartTime   string     `json:"startTime"`
	EndTime     string     `json:"endTime"`
	Status      string     `json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority    string     `json:"priority" enums:"High,Medium,Low"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func (input CreateTaskRequest) ToTaskInput() models.TaskInput {
	return models.TaskInput{
		Title:       input.Title,
		Subject:     input.Subject,
		Description: input.Description,
		Date:        input.Date,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      input.Status,
		Priority:    input.Priority,
	}
}

func (input UpdateTaskRequest) ToTaskInput() models.TaskInput {
	return CreateTaskRequest(input).ToTaskInput()
}

func (query TaskListQuery) ToTaskFilter() models.TaskFilter {
	return models.TaskFilter{
		Search:   query.Search,
		Subject:  query.Subject,
		Status:   query.Status,
		Priority: query.Priority,
		Date:     query.Date,
		From:     query.From,
		To:       query.To,
		View:     query.View,
		Sort:     query.Sort,
	}
}

func TaskResponseFromModel(task models.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Subject:     task.Subject,
		Description: task.Description,
		Date:        task.Date,
		StartTime:   task.StartTime,
		EndTime:     task.EndTime,
		Status:      task.Status,
		Priority:    task.Priority,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
	}
}

func TaskResponsesFromModels(tasks []models.Task) []TaskResponse {
	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, TaskResponseFromModel(task))
	}
	return responses
}
