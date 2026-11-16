package controllers

import (
	"errors"
	"net/http"

	"study-planner/dto"
	"study-planner/middleware"
	"study-planner/services"
	"study-planner/utils"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(service *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// Dashboard summarizes the authenticated user's tasks and weekly progress.
// @Summary Get dashboard
// @Tags dashboard
// @Produce json
// @Success 200 {object} dto.DashboardResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/dashboard [get]
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(r.Context(), middleware.UserID(r))
	if errors.Is(err, services.ErrLoadDashboardTasks) {
		utils.JSONError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}
	if errors.Is(err, services.ErrLoadStudyHistory) {
		utils.JSONError(w, http.StatusInternalServerError, "Could not load study completion history")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Could not load dashboard")
		return
	}
	utils.JSONResponse(w, http.StatusOK, dashboardResponseFromData(result))
}

func dashboardResponseFromData(data services.DashboardData) dto.DashboardResponse {
	week := make([]dto.DailyProgressResponse, 0, len(data.Week))
	for _, day := range data.Week {
		week = append(week, dto.DailyProgressResponse{
			Day:       day.Day,
			Date:      day.Date,
			Completed: day.Completed,
			Total:     day.Total,
			Progress:  day.Progress,
		})
	}

	return dto.DashboardResponse{
		TodayDate:      data.TodayDate,
		TodayProgress:  data.TodayProgress,
		TodayPlanned:   data.TodayPlanned,
		TodayCompleted: data.TodayCompleted,
		TodayRemaining: data.TodayRemaining,
		TotalTasks:     data.TotalTasks,
		Completed:      data.Completed,
		Pending:        data.Pending,
		NotStarted:     data.NotStarted,
		InProgress:     data.InProgress,
		OnHold:         data.OnHold,
		WeeklyProgressResponse: dto.WeeklyProgressResponse{
			Week:                  week,
			WeeklyProgress:        data.WeeklyProgress,
			PreviousWeekProgress:  data.PreviousWeekProgress,
			PreviousWeekAvailable: data.PreviousWeekAvailable,
			WeekChange:            data.WeekChange,
		},
		StreakResponse: dto.StreakResponse{
			CurrentStreak: data.CurrentStreak,
			LongestStreak: data.LongestStreak,
			LastStudyDate: data.LastStudyDate,
		},
	}
}
