package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"study-planner/middleware"
	"study-planner/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type taskRequest struct {
	Title              string  `json:"title"`
	Subject            string  `json:"subject"`
	Description        string  `json:"description"`
	Date               string  `json:"date"`
	StartTime          string  `json:"startTime"`
	EndTime            string  `json:"endTime"`
	Status             string  `json:"status" enums:"Not Started,In Progress,Completed,On Hold"`
	Priority           *string `json:"priority" enums:"High,Medium,Low"`
	ActualStudyMinutes *int    `json:"actualStudyMinutes" minimum:"0" maximum:"1440"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func filteredTaskQuery(r *http.Request) (*gorm.DB, error) {
	query := DB.Where("user_id = ?", middleware.UserID(r))
	params := r.URL.Query()
	today := time.Now().Format("2006-01-02")

	if search := strings.TrimSpace(params.Get("q")); search != "" {
		like := "%" + search + "%"
		query = query.Where("(title LIKE ? OR description LIKE ?)", like, like)
	}
	if subject := strings.TrimSpace(params.Get("subject")); subject != "" {
		query = query.Where("subject = ?", subject)
	}
	if status := params.Get("status"); status != "" {
		if status == "Pending" {
			status = "Not Started"
		}
		if !validTaskStatus(status) {
			return nil, fmt.Errorf("Invalid status filter")
		}
		query = query.Where("status = ?", status)
	}
	if priority := params.Get("priority"); priority != "" {
		if priority != "High" && priority != "Medium" && priority != "Low" {
			return nil, fmt.Errorf("Invalid priority filter")
		}
		query = query.Where("priority = ?", priority)
	}
	for _, key := range []string{"date", "from", "to"} {
		value := params.Get(key)
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return nil, fmt.Errorf("%s must use YYYY-MM-DD", key)
		}
		switch key {
		case "date":
			query = query.Where("date = ?", value)
		case "from":
			query = query.Where("date >= ?", value)
		case "to":
			query = query.Where("date <= ?", value)
		}
	}
	if from, to := params.Get("from"), params.Get("to"); from != "" && to != "" && from > to {
		return nil, fmt.Errorf("from must be on or before to")
	}

	switch params.Get("view") {
	case "", "all":
	case "completed":
		query = query.Where("status = ?", "Completed")
	case "pending":
		query = query.Where("status <> ?", "Completed")
	case "overdue":
		query = query.Where("date IS NOT NULL AND date <> '' AND date < ? AND status <> ?", today, "Completed")
	case "upcoming":
		query = query.Where("date >= ? AND status <> ?", today, "Completed")
	default:
		return nil, fmt.Errorf("Invalid task view")
	}

	switch params.Get("sort") {
	case "", "default", "dueDate":
		query = query.
			Order("CASE WHEN status = 'Completed' THEN 1 ELSE 0 END ASC").
			Order("CASE WHEN date IS NULL OR date = '' THEN 1 ELSE 0 END ASC").
			Order("CASE WHEN status <> 'Completed' AND date < '" + today + "' THEN 0 ELSE 1 END ASC").
			Order("CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END ASC").
			Order("date ASC, CASE WHEN start_time IS NULL OR start_time = '' THEN 1 ELSE 0 END ASC, start_time ASC, id ASC")
	case "priority":
		query = query.
			Order("CASE WHEN status = 'Completed' THEN 1 ELSE 0 END ASC").
			Order("CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END ASC").
			Order("CASE WHEN date IS NULL OR date = '' THEN 1 ELSE 0 END ASC, date ASC, CASE WHEN start_time IS NULL OR start_time = '' THEN 1 ELSE 0 END ASC, start_time ASC, id ASC")
	case "newest":
		query = query.Order("created_at DESC, id DESC")
	case "oldest":
		query = query.Order("created_at ASC, id ASC")
	case "status":
		query = query.Order("status ASC, date ASC, start_time ASC, id ASC")
	default:
		return nil, fmt.Errorf("Invalid sort option")
	}
	return query, nil
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
	if input.Status == "" {
		input.Status = "Not Started"
	}
	if input.Status == "Pending" {
		input.Status = "Not Started"
	}
	if input.Status != "Not Started" && input.Status != "In Progress" && input.Status != "Completed" && input.Status != "On Hold" {
		return "Status must be Not Started, In Progress, Completed, or On Hold"
	}
	if input.Priority == nil || *input.Priority == "" {
		defaultPriority := "Medium"
		input.Priority = &defaultPriority
	}
	if *input.Priority != "High" && *input.Priority != "Medium" && *input.Priority != "Low" {
		return "Priority must be High, Medium, or Low"
	}
	if input.ActualStudyMinutes != nil && (*input.ActualStudyMinutes < 0 || *input.ActualStudyMinutes > 1440) {
		return "Actual study minutes must be between 0 and 1440"
	}
	return ""
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
func CreateTask(w http.ResponseWriter, r *http.Request) {
	var input taskRequest
	if !DecodeJSON(w, r, &input) {
		return
	}
	if message := validateTask(&input); message != "" {
		JSONError(w, http.StatusBadRequest, message)
		return
	}

	task := models.Task{
		UserID:      middleware.UserID(r),
		Title:       input.Title,
		Subject:     input.Subject,
		Description: input.Description,
		Date:        input.Date,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      input.Status,
		Priority:    *input.Priority,
	}
	if input.ActualStudyMinutes != nil {
		task.ActualStudyMinutes = *input.ActualStudyMinutes
	}
	if task.Status == "Completed" {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		if task.Status == "Completed" {
			return tx.Create(&models.StudyCompletion{
				UserID:             task.UserID,
				TaskID:             task.ID,
				ActualStudyMinutes: task.ActualStudyMinutes,
				CompletedAt:        *task.CompletedAt,
			}).Error
		}
		return nil
	}); err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not create task")
		return
	}
	JSONResponse(w, http.StatusCreated, task)
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
func GetTasks(w http.ResponseWriter, r *http.Request) {
	query, err := filteredTaskQuery(r)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	var tasks []models.Task
	if err := query.Find(&tasks).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not get tasks")
		return
	}
	w.Header().Set("X-Server-Date", time.Now().Format("2006-01-02"))
	JSONResponse(w, http.StatusOK, tasks)
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
func GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var task models.Task
	if err := DB.Where("id = ? AND user_id = ?", id, middleware.UserID(r)).First(&task).Error; err != nil {
		writeTaskLookupError(w, err)
		return
	}
	JSONResponse(w, http.StatusOK, task)
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
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var input taskRequest
	if !DecodeJSON(w, r, &input) {
		return
	}
	priorityProvided := input.Priority != nil && *input.Priority != ""
	actualMinutesProvided := input.ActualStudyMinutes != nil
	if message := validateTask(&input); message != "" {
		JSONError(w, http.StatusBadRequest, message)
		return
	}

	var task models.Task
	if err := DB.Where("id = ? AND user_id = ?", id, middleware.UserID(r)).First(&task).Error; err != nil {
		writeTaskLookupError(w, err)
		return
	}
	wasCompleted := task.Status == "Completed"
	hadCompletionTimestamp := task.CompletedAt != nil
	task.Title = input.Title
	task.Subject = input.Subject
	task.Description = input.Description
	task.Date = input.Date
	task.StartTime = input.StartTime
	task.EndTime = input.EndTime
	task.Status = input.Status
	if priorityProvided {
		task.Priority = *input.Priority
	}
	if actualMinutesProvided {
		task.ActualStudyMinutes = *input.ActualStudyMinutes
	}
	if task.Status == "Completed" {
		if task.CompletedAt == nil {
			completedAt := time.Now()
			task.CompletedAt = &completedAt
		}
	} else {
		task.CompletedAt = nil
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if task.Status == "Completed" && (!wasCompleted || !hadCompletionTimestamp) {
			return tx.Create(&models.StudyCompletion{
				UserID:             task.UserID,
				TaskID:             task.ID,
				ActualStudyMinutes: task.ActualStudyMinutes,
				CompletedAt:        *task.CompletedAt,
			}).Error
		}
		if task.Status == "Completed" && actualMinutesProvided {
			var completion models.StudyCompletion
			err := tx.Where(
				"user_id = ? AND task_id = ? AND completed_at = ?",
				task.UserID, task.ID, *task.CompletedAt,
			).First(&completion).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(&models.StudyCompletion{
					UserID:             task.UserID,
					TaskID:             task.ID,
					ActualStudyMinutes: task.ActualStudyMinutes,
					CompletedAt:        *task.CompletedAt,
				}).Error
			}
			if err != nil {
				return err
			}
			return tx.Model(&completion).Update("actual_study_minutes", task.ActualStudyMinutes).Error
		}
		return nil
	}); err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not update task")
		return
	}
	JSONResponse(w, http.StatusOK, task)
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
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	result := DB.Where("id = ? AND user_id = ?", id, middleware.UserID(r)).Delete(&models.Task{})
	if result.Error != nil {
		JSONError(w, http.StatusInternalServerError, "Could not delete task")
		return
	}
	if result.RowsAffected == 0 {
		JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	JSONResponse(w, http.StatusOK, messageResponse{Message: "Task deleted"})
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
func CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var task models.Task
	if err := DB.Where("id = ? AND user_id = ?", id, middleware.UserID(r)).First(&task).Error; err != nil {
		writeTaskLookupError(w, err)
		return
	}
	wasCompleted := task.Status == "Completed" && task.CompletedAt != nil
	if !wasCompleted {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	task.Status = "Completed"
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if !wasCompleted {
			return tx.Create(&models.StudyCompletion{
				UserID:             task.UserID,
				TaskID:             task.ID,
				ActualStudyMinutes: task.ActualStudyMinutes,
				CompletedAt:        *task.CompletedAt,
			}).Error
		}
		return nil
	}); err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not complete task")
		return
	}

	JSONResponse(w, http.StatusOK, messageResponse{Message: "Task completed"})
}

func taskID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		JSONError(w, http.StatusBadRequest, "Invalid task ID")
		return 0, false
	}
	return uint(id), true
}

func writeTaskLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	JSONError(w, http.StatusInternalServerError, "Could not get task")
}
