package handlers

import (
	"net/http"
	"time"

	"study-planner/middleware"
	"study-planner/models"
)

type dayProgress struct {
	Day       string `json:"day"`
	Date      string `json:"date"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Progress  int    `json:"progress"`
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserID(r)
	var tasks []models.Task
	if err := DB.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}

	today := time.Now().Format("2006-01-02")
	total, completed, todayTotal, todayCompleted, studyMinutes := 0, 0, 0, 0, 0
	for _, task := range tasks {
		total++
		if task.Status == "Completed" {
			completed++
		}
		if task.Date == today {
			todayTotal++
			if task.Status == "Completed" {
				todayCompleted++
			}
			start, startErr := time.Parse("15:04", task.StartTime)
			end, endErr := time.Parse("15:04", task.EndTime)
			if startErr == nil && endErr == nil && end.After(start) {
				studyMinutes += int(end.Sub(start).Minutes())
			}
		}
	}

	todayProgress := 0
	if todayTotal > 0 {
		todayProgress = todayCompleted * 100 / todayTotal
	}
	pending := total - completed

	now := time.Now()
	weekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart = weekStart.AddDate(0, 0, -((int(weekStart.Weekday()) + 6) % 7))
	week := make([]dayProgress, 0, 7)
	for i := 0; i < 7; i++ {
		date := weekStart.AddDate(0, 0, i).Format("2006-01-02")
		day := dayProgress{Day: weekStart.AddDate(0, 0, i).Weekday().String(), Date: date}
		for _, task := range tasks {
			if task.Date == date {
				day.Total++
				if task.Status == "Completed" {
					day.Completed++
				}
			}
		}
		if day.Total > 0 {
			day.Progress = day.Completed * 100 / day.Total
		}
		week = append(week, day)
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"todayProgress": todayProgress,
		"totalTasks":    total,
		"completed":     completed,
		"pending":       pending,
		"studyMinutes":  studyMinutes,
		"week":          week,
	})
}
