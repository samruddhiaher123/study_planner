package controllers

import (
	"errors"
	"net/http"
	"time"

	"study-planner/dto"
	"study-planner/middleware"
	"study-planner/services"
	"study-planner/utils"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// CreateTask creates a task for the authenticated user.
// @Summary Create a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "Task details"
// @Success 201 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateTaskRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}
	if message := validateTask(&input); message != "" {
		utils.JSONError(w, http.StatusBadRequest, message)
		return
	}
	task, err := h.service.Create(r.Context(), middleware.UserID(r), input.ToTaskInput())
	if err != nil {
		if errors.Is(err, services.ErrInvalidTaskStatus) || errors.Is(err, services.ErrInvalidPriority) {
			utils.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.JSONError(w, http.StatusInternalServerError, "Could not create task")
		return
	}
	utils.JSONResponse(w, http.StatusCreated, dto.TaskResponseFromModel(task))
}

// GetTasks lists tasks belonging to the authenticated user.
// @Summary List tasks
// @Tags tasks
// @Produce json
// @Param q query string false "Search task titles and descriptions"
// @Param subject query string false "Filter by exact subject"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param date query string false "Filter by scheduled date (YYYY-MM-DD)"
// @Param from query string false "Start of scheduled date range (YYYY-MM-DD)"
// @Param to query string false "End of scheduled date range (YYYY-MM-DD)"
// @Param view query string false "Task view: pending, completed, overdue, or upcoming"
// @Param sort query string false "Sort: default, priority, dueDate, newest, oldest, or status"
// @Success 200 {array} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	filter, err := parseTaskFilter(r)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	tasks, err := h.service.List(r.Context(), middleware.UserID(r), filter)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Could not get tasks")
		return
	}
	w.Header().Set("X-Server-Date", time.Now().Format("2006-01-02"))
	utils.JSONResponse(w, http.StatusOK, dto.TaskResponsesFromModels(tasks))
}

// GetTask returns one task belonging to the authenticated user.
// @Summary Get a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	task, err := h.service.Get(r.Context(), id, middleware.UserID(r))
	if err != nil {
		if errors.Is(err, services.ErrTaskNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Task not found")
		} else {
			utils.JSONError(w, http.StatusInternalServerError, "Could not get task")
		}
		return
	}
	utils.JSONResponse(w, http.StatusOK, dto.TaskResponseFromModel(task))
}

// UpdateTask replaces the editable fields of a task belonging to the authenticated user.
// @Summary Update a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path uint true "Task ID"
// @Param request body dto.UpdateTaskRequest true "Updated task details"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var input dto.UpdateTaskRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}
	validatedInput := dto.CreateTaskRequest(input)
	if message := validateTask(&validatedInput); message != "" {
		utils.JSONError(w, http.StatusBadRequest, message)
		return
	}

	task, err := h.service.Update(
		r.Context(),
		id,
		middleware.UserID(r),
		validatedInput.ToTaskInput(),
	)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTaskStatus) || errors.Is(err, services.ErrInvalidPriority) {
			utils.JSONError(w, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, services.ErrTaskNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Task not found")
		} else {
			utils.JSONError(w, http.StatusInternalServerError, "Could not update task")
		}
		return
	}
	utils.JSONResponse(w, http.StatusOK, dto.TaskResponseFromModel(task))
}

// DeleteTask deletes a task belonging to the authenticated user.
// @Summary Delete a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id, middleware.UserID(r)); err != nil {
		if errors.Is(err, services.ErrTaskNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Task not found")
		} else {
			utils.JSONError(w, http.StatusInternalServerError, "Could not delete task")
		}
		return
	}
	utils.JSONResponse(w, http.StatusOK, dto.MessageResponse{Message: "Task deleted"})
}

// CompleteTask marks a task belonging to the authenticated user as completed.
// @Summary Complete a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /api/tasks/{id}/complete [put]
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	if err := h.service.Complete(r.Context(), id, middleware.UserID(r)); err != nil {
		if errors.Is(err, services.ErrTaskNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Task not found")
		} else {
			utils.JSONError(w, http.StatusInternalServerError, "Could not complete task")
		}
		return
	}
	utils.JSONResponse(w, http.StatusOK, dto.MessageResponse{Message: "Task completed"})
}
