package services

import (
	"context"
	"errors"
	"time"

	"study-planner/models"
	"study-planner/repositories"

	"gorm.io/gorm"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskService struct {
	tasks repositories.TaskRepository
}

func NewTaskService(tasks repositories.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func (s *TaskService) Create(ctx context.Context, userID uint, input models.TaskInput) (models.Task, error) {
	task := taskFromInput(userID, input)
	if task.Status == "Completed" {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	if err := s.tasks.Create(ctx, &task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (s *TaskService) List(ctx context.Context, userID uint, filter models.TaskFilter) ([]models.Task, error) {
	return s.tasks.List(ctx, userID, filter)
}

func (s *TaskService) Get(ctx context.Context, id, userID uint) (models.Task, error) {
	task, err := s.tasks.GetByID(ctx, id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Task{}, ErrTaskNotFound
	}
	return task, err
}

func (s *TaskService) Update(ctx context.Context, id, userID uint, input models.TaskInput) (models.Task, error) {
	task, err := s.Get(ctx, id, userID)
	if err != nil {
		return models.Task{}, err
	}

	wasCompleted := task.Status == "Completed"
	hadCompletionTimestamp := task.CompletedAt != nil
	task.Title = input.Title
	task.Subject = input.Subject
	task.Description = input.Description
	task.Date = input.Date
	task.StartTime = input.StartTime
	task.EndTime = input.EndTime
	task.Status = input.Status
	if input.Priority != nil && *input.Priority != "" {
		task.Priority = *input.Priority
	}
	if task.Status == "Completed" && task.CompletedAt == nil {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	} else if task.Status != "Completed" {
		task.CompletedAt = nil
	}

	addCompletion := task.Status == "Completed" && (!wasCompleted || !hadCompletionTimestamp)

	if err := s.tasks.Update(ctx, &task, addCompletion); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (s *TaskService) Complete(ctx context.Context, id, userID uint) error {
	task, err := s.Get(ctx, id, userID)
	if err != nil {
		return err
	}
	//task completed if status is "Completed" and CompletedAt is not nil

	wasCompleted := task.Status == "Completed" && task.CompletedAt != nil
	if !wasCompleted {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	task.Status = "Completed"

	return s.tasks.Update(ctx, &task, !wasCompleted)
}

func (s *TaskService) Delete(ctx context.Context, id, userID uint) error {
	deleted, err := s.tasks.Delete(ctx, id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrTaskNotFound
	}
	return nil
}

func taskFromInput(userID uint, input models.TaskInput) models.Task {
	priority := "Medium"
	if input.Priority != nil && *input.Priority != "" {
		priority = *input.Priority
	}
	return models.Task{
		UserID:      userID,
		Title:       input.Title,
		Subject:     input.Subject,
		Description: input.Description,
		Date:        input.Date,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      input.Status,
		Priority:    priority,
	}
}
