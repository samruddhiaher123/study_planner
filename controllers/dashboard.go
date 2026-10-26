package controllers

import (
	"errors"
	"net/http"

	"study-planner/middleware"
	"study-planner/models"
	"study-planner/services"
	"study-planner/utils"
)

type DashboardResponse = models.DashboardResponse

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
// @Success 200 {object} DashboardResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
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
	utils.JSONResponse(w, http.StatusOK, result)
}
