package handlers

import (
	"net/http"
	"sort"
	"time"

	"study-planner/middleware"
	"study-planner/models"
)

type DayProgress struct {
	Day          string `json:"day"`
	Date         string `json:"date"`
	Completed    int    `json:"completed"`
	Total        int    `json:"total"`
	Progress     int    `json:"progress"`
	StudyMinutes int    `json:"studyMinutes"`
}

type DashboardResponse struct {
	TodayDate             string        `json:"todayDate"`
	TodayProgress         int           `json:"todayProgress"`
	TodayPlanned          int           `json:"todayPlanned"`
	TodayCompleted        int           `json:"todayCompleted"`
	TodayRemaining        int           `json:"todayRemaining"`
	TotalTasks            int           `json:"totalTasks"`
	Completed             int           `json:"completed"`
	Pending               int           `json:"pending"`
	NotStarted            int           `json:"notStarted"`
	InProgress            int           `json:"inProgress"`
	OnHold                int           `json:"onHold"`
	StudyMinutes          int           `json:"studyMinutes"`
	Week                  []DayProgress `json:"week"`
	WeeklyProgress        int           `json:"weeklyProgress"`
	PreviousWeekProgress  int           `json:"previousWeekProgress"`
	PreviousWeekAvailable bool          `json:"previousWeekAvailable"`
	WeekChange            int           `json:"weekChange"`
	CurrentStreak         int           `json:"currentStreak"`
	LongestStreak         int           `json:"longestStreak"`
	LastStudyDate         string        `json:"lastStudyDate"`
}

// Dashboard summarizes the authenticated user's tasks and weekly progress.
// @Summary Get dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} DashboardResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Security BearerAuth
// @Router /api/dashboard [get]
func Dashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	var tasks []models.Task
	if err := DB.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}
	var completions []models.StudyCompletion
	if err := DB.Where("user_id = ?", userID).Find(&completions).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not load study completion history")
		return
	}

	now := time.Now()
	today := dateOnly(now)
	todayString := today.Format("2006-01-02")
	weekStart := startOfWeek(today)
	previousWeekStart := weekStart.AddDate(0, 0, -7)

	result := DashboardResponse{TodayDate: todayString, Week: make([]DayProgress, 7)}
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
		day := DayProgress{Day: dayDate.Weekday().String(), Date: dayDate.Format("2006-01-02")}
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
		completedOn := dateOnly(completion.CompletedAt.In(today.Location()))
		if completedOn.Equal(today) {
			result.StudyMinutes += completion.ActualStudyMinutes
		}
		if !completedOn.Before(weekStart) && completedOn.Before(weekStart.AddDate(0, 0, 7)) {
			result.Week[calendarDateOffset(completedOn, weekStart)].StudyMinutes += completion.ActualStudyMinutes
		}
		completionDates = append(completionDates, completion.CompletedAt)
	}
	if weeklyTotal > 0 {
		result.WeeklyProgress = weeklyCompleted * 100 / weeklyTotal
	}
	if result.PreviousWeekAvailable {
		result.WeekChange = result.WeeklyProgress - result.PreviousWeekProgress
	}

	result.CurrentStreak, result.LongestStreak, result.LastStudyDate = calculateStreak(completionDates, today)
	JSONResponse(w, http.StatusOK, result)
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func startOfWeek(day time.Time) time.Time {
	return day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
}

func calendarDateOffset(day, start time.Time) int {
	y, m, d := day.Date()
	dayUTC := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = start.Date()
	startUTC := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return int(dayUTC.Sub(startUTC).Hours() / 24)
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
