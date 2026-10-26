package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"study-planner/models"
	"study-planner/repositories"
)

var (
	ErrLoadDashboardTasks = errors.New("could not load dashboard")
	ErrLoadStudyHistory   = errors.New("could not load study completion history")
)

type DashboardService struct {
	tasks repositories.TaskRepository
}

func NewDashboardService(tasks repositories.TaskRepository) *DashboardService {
	return &DashboardService{tasks: tasks}
}

func (s *DashboardService) Get(ctx context.Context, userID uint) (models.DashboardResponse, error) {
	tasks, err := s.tasks.ListForDashboard(ctx, userID)
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("%w: %v", ErrLoadDashboardTasks, err)
	}
	completions, err := s.tasks.CompletionHistory(ctx, userID)
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("%w: %v", ErrLoadStudyHistory, err)
	}

	now := time.Now()
	today := dateOnly(now)
	todayString := today.Format("2006-01-02")
	weekStart := startOfWeek(today)
	previousWeekStart := weekStart.AddDate(0, 0, -7)
	result := models.DashboardResponse{TodayDate: todayString, Week: make([]models.DayProgress, 7)}

	for _, task := range tasks {
		result.TotalTasks++
		switch task.Status {
		case "Completed":
			result.Completed++
		case "In Progress":
			result.InProgress++
		case "On Hold":
			result.OnHold++
		default:
			result.NotStarted++
		}
		if task.Date == todayString {
			result.TodayPlanned++
			if task.Status == "Completed" {
				result.TodayCompleted++
			}
		}
		scheduledDate, err := time.ParseInLocation("2006-01-02", task.Date, today.Location())
		if err == nil && !scheduledDate.Before(previousWeekStart) && scheduledDate.Before(weekStart) {
			result.PreviousWeekAvailable = true
			if task.Status == "Completed" {
				result.PreviousWeekProgress++
			}
		}
	}

	result.Pending = result.TotalTasks - result.Completed
	result.TodayRemaining = result.TodayPlanned - result.TodayCompleted
	if result.TodayPlanned > 0 {
		result.TodayProgress = result.TodayCompleted * 100 / result.TodayPlanned
	}
	if result.PreviousWeekAvailable {
		result.PreviousWeekProgress = result.PreviousWeekProgress * 100 / countTasksInRange(tasks, previousWeekStart, weekStart)
	}

	weeklyTotal, weeklyCompleted := 0, 0
	for i := 0; i < 7; i++ {
		dayDate := weekStart.AddDate(0, 0, i)
		day := models.DayProgress{Day: dayDate.Weekday().String(), Date: dayDate.Format("2006-01-02")}
		for _, task := range tasks {
			if task.Date == day.Date {
				day.Total++
				if task.Status == "Completed" {
					day.Completed++
				}
			}
		}
		if day.Total > 0 {
			day.Progress = day.Completed * 100 / day.Total
		}
		weeklyTotal += day.Total
		weeklyCompleted += day.Completed
		result.Week[i] = day
	}

	completionDates := make([]time.Time, 0, len(completions))
	for _, completion := range completions {
		completionDates = append(completionDates, completion.CompletedAt)
	}
	if weeklyTotal > 0 {
		result.WeeklyProgress = weeklyCompleted * 100 / weeklyTotal
	}
	if result.PreviousWeekAvailable {
		result.WeekChange = result.WeeklyProgress - result.PreviousWeekProgress
	}
	result.CurrentStreak, result.LongestStreak, result.LastStudyDate = calculateStreak(completionDates, today)
	return result, nil
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func startOfWeek(day time.Time) time.Time {
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

func countTasksInRange(tasks []models.Task, from, to time.Time) int {
	total := 0
	for _, task := range tasks {
		day, err := time.ParseInLocation("2006-01-02", task.Date, from.Location())
		if err == nil && !day.Before(from) && day.Before(to) {
			total++
		}
	}
	return total
}

func calculateStreak(completedAt []time.Time, today time.Time) (current, longest int, lastStudyDate string) {
	days := make(map[string]struct{}, len(completedAt))
	latest := time.Time{}
	for _, completed := range completedAt {
		day := dateOnly(completed.In(today.Location()))
		if day.After(today) {
			continue
		}
		key := day.Format("2006-01-02")
		days[key] = struct{}{}
		if day.After(latest) {
			latest = day
		}
	}
	if !latest.IsZero() {
		lastStudyDate = latest.Format("2006-01-02")
	}

	ordered := make([]time.Time, 0, len(days))
	for key := range days {
		day, _ := time.ParseInLocation("2006-01-02", key, today.Location())
		ordered = append(ordered, day)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Before(ordered[j]) })
	for i, day := range ordered {
		if i == 0 || !day.Equal(ordered[i-1].AddDate(0, 0, 1)) {
			currentRun := 1
			for j := i + 1; j < len(ordered) && ordered[j].Equal(ordered[j-1].AddDate(0, 0, 1)); j++ {
				currentRun++
			}
			if currentRun > longest {
				longest = currentRun
			}
		}
	}

	anchor := today
	if _, completedToday := days[anchor.Format("2006-01-02")]; !completedToday {
		anchor = today.AddDate(0, 0, -1)
	}
	for {
		if _, ok := days[anchor.Format("2006-01-02")]; !ok {
			break
		}
		current++
		anchor = anchor.AddDate(0, 0, -1)
	}
	return current, longest, lastStudyDate
}
