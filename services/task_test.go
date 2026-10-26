package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"study-planner/models"

	"gorm.io/gorm"
)

type taskRepositoryStub struct {
	task            models.Task
	addedCompletion bool
	updateCalls     int
}

func (r *taskRepositoryStub) Create(_ context.Context, task *models.Task) error {
	r.task = *task
	return nil
}

func (r *taskRepositoryStub) List(context.Context, uint, models.TaskFilter) ([]models.Task, error) {
	return nil, nil
}

func (r *taskRepositoryStub) GetByID(_ context.Context, id, userID uint) (models.Task, error) {
	if r.task.ID != id || r.task.UserID != userID {
		return models.Task{}, gorm.ErrRecordNotFound
	}
	return r.task, nil
}

func (r *taskRepositoryStub) Update(_ context.Context, task *models.Task, addCompletion bool) error {
	r.task = *task
	r.addedCompletion = addCompletion
	r.updateCalls++
	return nil
}

func (r *taskRepositoryStub) Delete(_ context.Context, id, userID uint) (bool, error) {
	if r.task.ID != id || r.task.UserID != userID {
		return false, nil
	}
	return true, nil
}

func (r *taskRepositoryStub) ListForDashboard(context.Context, uint) ([]models.Task, error) {
	return []models.Task{r.task}, nil
}

func (r *taskRepositoryStub) CompletionHistory(context.Context, uint) ([]models.StudyCompletion, error) {
	return nil, nil
}

func TestUpdateCompletionAndOptionalPriority(t *testing.T) {
	repository := &taskRepositoryStub{
		task: models.Task{
			ID:       7,
			UserID:   12,
			Title:    "Review notes",
			Subject:  "Biology",
			Status:   "Not Started",
			Priority: "High",
		},
	}
	service := NewTaskService(repository)
	input := models.TaskInput{
		Title:     "Review notes",
		Subject:   "Biology",
		Date:      "2026-10-26",
		StartTime: "09:00",
		EndTime:   "10:00",
		Status:    "Completed",
	}

	updated, err := service.Update(context.Background(), 7, 12, input)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Priority != "High" {
		t.Errorf("Update() priority = %q, want existing priority High", updated.Priority)
	}
	if updated.CompletedAt == nil {
		t.Fatal("Update() did not set the completion timestamp")
	}
	if !repository.addedCompletion {
		t.Error("Update() did not request a completion-history record")
	}

	if err := service.Complete(context.Background(), 7, 12); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if repository.addedCompletion {
		t.Error("Complete() added duplicate history for an already completed task")
	}
	if repository.updateCalls != 2 {
		t.Errorf("repository update calls = %d, want 2", repository.updateCalls)
	}
}

func TestCreateDefaultsMissingPriority(t *testing.T) {
	repository := &taskRepositoryStub{}
	service := NewTaskService(repository)

	created, err := service.Create(context.Background(), 12, models.TaskInput{
		Title:     "Review notes",
		Subject:   "Biology",
		Date:      "2026-10-26",
		StartTime: "09:00",
		EndTime:   "10:00",
		Status:    "Not Started",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Priority != "Medium" {
		t.Errorf("Create() priority = %q, want Medium", created.Priority)
	}
}

func TestUpdateLeavingCompletedClearsCurrentTimestamp(t *testing.T) {
	completedAt := time.Now()
	repository := &taskRepositoryStub{
		task: models.Task{
			ID:          7,
			UserID:      12,
			Status:      "Completed",
			Priority:    "Medium",
			CompletedAt: &completedAt,
		},
	}
	service := NewTaskService(repository)

	updated, err := service.Update(context.Background(), 7, 12, models.TaskInput{
		Title:     "Review notes",
		Subject:   "Biology",
		Date:      "2026-10-26",
		StartTime: "09:00",
		EndTime:   "10:00",
		Status:    "On Hold",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.CompletedAt != nil {
		t.Error("Update() retained completion timestamp after leaving Completed")
	}
	if repository.addedCompletion {
		t.Error("Update() added completion history when leaving Completed")
	}
}

func TestTaskAccessIsScopedToUser(t *testing.T) {
	repository := &taskRepositoryStub{task: models.Task{ID: 7, UserID: 12}}
	service := NewTaskService(repository)

	_, err := service.Get(context.Background(), 7, 99)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Get() error = %v, want ErrTaskNotFound for another user", err)
	}
}
