package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"study-planner/middleware"
	"study-planner/models"
	"study-planner/services"
	"study-planner/utils"

	"github.com/gorilla/mux"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

type taskRequest struct {
	Title       string  `json:"title"`
	Subject     string  `json:"subject"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	Status      string  `json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority    *string `json:"priority" enums:"High,Medium,Low"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// CreateTask creates a task for the authenticated user.
// @Summary Create a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body taskRequest true "Task details"
// @Success 201 {object} models.Task
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Security BearerAuth
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input taskRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}
	if message := validateTask(&input); message != "" {
		utils.JSONError(w, http.StatusBadRequest, message)
		return
	}
	task, err := h.service.Create(r.Context(), middleware.UserID(r), toTaskInput(input))
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Could not create task")
		return
	}
	utils.JSONResponse(w, http.StatusCreated, task)
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
// @Success 200 {array} models.Task
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
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
	if tasks == nil {
		tasks = make([]models.Task, 0)
	}
	w.Header().Set("X-Server-Date", time.Now().Format("2006-01-02"))
	utils.JSONResponse(w, http.StatusOK, tasks)
}

// GetTask returns one task belonging to the authenticated user.
// @Summary Get a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
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
	utils.JSONResponse(w, http.StatusOK, task)
}

// UpdateTask replaces the editable fields of a task belonging to the authenticated user.
// @Summary Update a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path uint true "Task ID"
// @Param request body taskRequest true "Updated task details"
// @Success 200 {object} models.Task
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Security BearerAuth
// @Router /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var input taskRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}
	if message := validateTask(&input); message != "" {
		utils.JSONError(w, http.StatusBadRequest, message)
		return
	}

	task, err := h.service.Update(
		r.Context(),
		id,
		middleware.UserID(r),
		toTaskInput(input),
	)
	if err != nil {
		if errors.Is(err, services.ErrTaskNotFound) {
			utils.JSONError(w, http.StatusNotFound, "Task not found")
		} else {
			utils.JSONError(w, http.StatusInternalServerError, "Could not update task")
		}
		return
	}
	utils.JSONResponse(w, http.StatusOK, task)
}

// DeleteTask deletes a task belonging to the authenticated user.
// @Summary Delete a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} messageResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
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
	utils.JSONResponse(w, http.StatusOK, messageResponse{Message: "Task deleted"})
}

// CompleteTask marks a task belonging to the authenticated user as completed.
// @Summary Complete a task
// @Tags tasks
// @Produce json
// @Param id path uint true "Task ID"
// @Success 200 {object} messageResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
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
	utils.JSONResponse(w, http.StatusOK, messageResponse{Message: "Task completed"})
}

func toTaskInput(input taskRequest) models.TaskInput {
	return models.TaskInput{
		Title:       input.Title,
		Subject:     input.Subject,
		Description: input.Description,
		Date:        input.Date,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      input.Status,
		Priority:    input.Priority,
	}
}

func taskID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid task ID")
		return 0, false
	}
	return uint(id), true
}

func writeTaskLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, services.ErrTaskNotFound) {
		utils.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	utils.JSONError(w, http.StatusInternalServerError, "Could not get task")
}

func parseTaskFilter(r *http.Request) (models.TaskFilter, error) {
	params := r.URL.Query()
	filter := models.TaskFilter{
		Search:   strings.TrimSpace(params.Get("q")),
		Subject:  strings.TrimSpace(params.Get("subject")),
		Status:   params.Get("status"),
		Priority: params.Get("priority"),
		Date:     params.Get("date"),
		From:     params.Get("from"),
		To:       params.Get("to"),
		View:     params.Get("view"),
		Sort:     params.Get("sort"),
	}
	if filter.Status == "Pending" {
		filter.Status = "Not Started"
	}
	if filter.Status != "" && !validTaskStatus(filter.Status) {
		return models.TaskFilter{}, fmt.Errorf("Invalid status filter")
	}
	if filter.Priority != "" && filter.Priority != "High" && filter.Priority != "Medium" && filter.Priority != "Low" {
		return models.TaskFilter{}, fmt.Errorf("Invalid priority filter")
	}
	for _, key := range []string{"date", "from", "to"} {
		value := params.Get(key)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return models.TaskFilter{}, fmt.Errorf("%s must use YYYY-MM-DD", key)
		}
	}
	if filter.From != "" && filter.To != "" && filter.From > filter.To {
		return models.TaskFilter{}, fmt.Errorf("from must be on or before to")
	}
	switch filter.View {
	case "", "all", "completed", "pending", "overdue", "upcoming":
	default:
		return models.TaskFilter{}, fmt.Errorf("Invalid task view")
	}
	switch filter.Sort {
	case "", "default", "dueDate", "priority", "newest", "oldest", "status":
	default:
		return models.TaskFilter{}, fmt.Errorf("Invalid sort option")
	}
	return filter, nil
}

func validTaskStatus(status string) bool {
	return status == "Not Started" || status == "In Progress" || status == "Completed" || status == "On Hold"
}

func validateTask(input *taskRequest) string {
	input.Title = strings.TrimSpace(input.Title)
	input.Subject = strings.TrimSpace(input.Subject)
	if input.Title == "" || input.Subject == "" {
		return "Title and subject are required"
	}
	if len([]rune(input.Title)) > 255 || len([]rune(input.Subject)) > 150 {
		return "Title or subject is too long"
	}
	if _, err := time.Parse("2006-01-02", input.Date); err != nil {
		return "Date must use YYYY-MM-DD"
	}
	start, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		return "Start time must use HH:MM"
	}
	end, err := time.Parse("15:04", input.EndTime)
	if err != nil || !end.After(start) {
		return "End time must use HH:MM and be later than start time"
	}
	if input.Status == "" || input.Status == "Pending" {
		input.Status = "Not Started"
	}
	if !validTaskStatus(input.Status) {
		return "Status must be Not Started, In Progress, Completed, or On Hold"
	}
	if input.Priority == nil || *input.Priority == "" {
		defaultPriority := "Medium"
		input.Priority = &defaultPriority
	}
	if *input.Priority != "High" && *input.Priority != "Medium" && *input.Priority != "Low" {
		return "Priority must be High, Medium, or Low"
	}
	return ""
}
