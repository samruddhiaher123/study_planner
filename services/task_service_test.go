package services

import (
	"context"
	"errors"
	"testing"

	"study-planner/models"
)

type taskServiceTestRepository struct {
	createdTask *models.Task
	createCalls int
	task        models.Task
	getCalls    int
	updateCalls int
}

func (r *taskServiceTestRepository) Create(_ context.Context, task *models.Task) error {
	r.createdTask = task
	r.createCalls++
	return nil
}

func (r *taskServiceTestRepository) List(context.Context, uint, models.TaskFilter) ([]models.Task, error) {
	return nil, nil
}

func (r *taskServiceTestRepository) GetByID(context.Context, uint, uint) (models.Task, error) {
	r.getCalls++
	if r.task.ID == 0 {
		return models.Task{}, errors.New("test task not found")
	}
	return r.task, nil
}

func (r *taskServiceTestRepository) Update(_ context.Context, task *models.Task, _ bool) error {
	r.task = *task
	r.updateCalls++
	return nil
}

func (r *taskServiceTestRepository) Delete(context.Context, uint, uint) (bool, error) {
	return false, nil
}

func (r *taskServiceTestRepository) ListForDashboard(context.Context, uint) ([]models.Task, error) {
	return nil, nil
}

func (r *taskServiceTestRepository) CompletionHistory(context.Context, uint) ([]models.StudyCompletion, error) {
	return nil, nil
}

func TestTaskServiceCreateValidatesStatusAndPriority(t *testing.T) {
	tests := []struct {
		name    string
		input   models.TaskInput
		wantErr error
	}{
		{
			name: "invalid status",
			input: models.TaskInput{
				Status:   "Invalid",
				Priority: taskPriorityPointer(models.PriorityHigh),
			},
			wantErr: ErrInvalidTaskStatus,
		},
		{
			name: "invalid priority",
			input: models.TaskInput{
				Status:   models.TaskNotStarted,
				Priority: taskPriorityPointer("Urgent"),
			},
			wantErr: ErrInvalidPriority,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &taskServiceTestRepository{}
			service := NewTaskService(repository)

			_, err := service.Create(context.Background(), 1, test.input)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, test.wantErr)
			}
			if repository.createCalls != 0 {
				t.Fatalf("Create() wrote to repository %d times for invalid input", repository.createCalls)
			}
		})
	}
}

func TestTaskServiceCreateUsesDefaults(t *testing.T) {
	repository := &taskServiceTestRepository{}
	service := NewTaskService(repository)

	task, err := service.Create(context.Background(), 1, models.TaskInput{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.Status != models.TaskNotStarted {
		t.Errorf("Create() status = %q, want %q", task.Status, models.TaskNotStarted)
	}
	if task.Priority != models.PriorityMedium {
		t.Errorf("Create() priority = %q, want %q", task.Priority, models.PriorityMedium)
	}
	if repository.createCalls != 1 {
		t.Errorf("repository Create() calls = %d, want 1", repository.createCalls)
	}
}

func TestTaskServiceUpdateRejectsInvalidValuesBeforeReadingTask(t *testing.T) {
	tests := []struct {
		name    string
		input   models.TaskInput
		wantErr error
	}{
		{
			name: "invalid status",
			input: models.TaskInput{
				Status:   "Invalid",
				Priority: taskPriorityPointer(models.PriorityHigh),
			},
			wantErr: ErrInvalidTaskStatus,
		},
		{
			name: "invalid priority",
			input: models.TaskInput{
				Status:   models.TaskNotStarted,
				Priority: taskPriorityPointer("Urgent"),
			},
			wantErr: ErrInvalidPriority,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &taskServiceTestRepository{}
			service := NewTaskService(repository)

			_, err := service.Update(context.Background(), 1, 1, test.input)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Update() error = %v, want %v", err, test.wantErr)
			}
			if repository.getCalls != 0 || repository.updateCalls != 0 {
				t.Fatalf("Update() accessed repository for invalid input: get=%d update=%d", repository.getCalls, repository.updateCalls)
			}
		})
	}
}

func taskPriorityPointer(priority models.TaskPriority) *models.TaskPriority {
	return &priority
}
