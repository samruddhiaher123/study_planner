package services

import (
	"context"
	"testing"
	"time"

	"study-planner/models"
)

type dashboardRepositoryStub struct {
	tasks      []models.Task
	completion []models.StudyCompletion
}

func (r *dashboardRepositoryStub) Create(context.Context, *models.Task) error {
	panic("unexpected Create call")
}

func (r *dashboardRepositoryStub) List(context.Context, uint, models.TaskFilter) ([]models.Task, error) {
	panic("unexpected List call")
}

func (r *dashboardRepositoryStub) GetByID(context.Context, uint, uint) (models.Task, error) {
	panic("unexpected GetByID call")
}

func (r *dashboardRepositoryStub) Update(context.Context, *models.Task, bool) error {
	panic("unexpected Update call")
}

func (r *dashboardRepositoryStub) Delete(context.Context, uint, uint) (bool, error) {
	panic("unexpected Delete call")
}

func (r *dashboardRepositoryStub) ListForDashboard(context.Context, uint) ([]models.Task, error) {
	return r.tasks, nil
}

func (r *dashboardRepositoryStub) CompletionHistory(context.Context, uint) ([]models.StudyCompletion, error) {
	return r.completion, nil
}

func TestDashboardProgressAndStreakUsePersistedCompletions(t *testing.T) {
	now := time.Now()
	today := dateOnly(now).Format("2006-01-02")
	yesterday := dateOnly(now).AddDate(0, 0, -1)
	repository := &dashboardRepositoryStub{
		tasks: []models.Task{
			{Date: today, Status: "Completed"},
			{Date: today, Status: "In Progress"},
		},
		completion: []models.StudyCompletion{
			{CompletedAt: yesterday.Add(12 * time.Hour)},
			{CompletedAt: now},
		},
	}
	service := NewDashboardService(repository)

	result, err := service.Get(context.Background(), 11)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result.TodayPlanned != 2 || result.TodayCompleted != 1 || result.TodayProgress != 50 {
		t.Errorf("today progress = planned %d, completed %d, percent %d; want 2, 1, 50",
			result.TodayPlanned, result.TodayCompleted, result.TodayProgress)
	}
	if result.CurrentStreak != 2 || result.LongestStreak != 2 || result.LastStudyDate != today {
		t.Errorf("streak = current %d, longest %d, last %q; want 2, 2, %s",
			result.CurrentStreak, result.LongestStreak, result.LastStudyDate, today)
	}
}
