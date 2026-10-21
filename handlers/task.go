package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"study-planner/middleware"
	"study-planner/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

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

	err := DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		if task.CompletedAt != nil {
			return tx.Create(&models.StudyCompletion{
				UserID:             task.UserID,
				TaskID:             task.ID,
				ActualStudyMinutes: task.ActualStudyMinutes,
				CompletedAt:        *task.CompletedAt,
			}).Error
		}
		return nil
	})
	if err != nil {
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
	query, args, err := filteredTaskQuery(r)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	tasks := make([]models.Task, 0)
	if err := DB.WithContext(r.Context()).Raw(query, args...).Scan(&tasks).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not get tasks")
		return
	}
	if tasks == nil {
		tasks = make([]models.Task, 0)
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
	task, err := getTask(r, id, middleware.UserID(r))
	if err != nil {
		writeTaskLookupError(w, err)
		return
	}
	JSONResponse(w, http.StatusOK, task)
}

func getTask(r *http.Request, id, userID uint) (models.Task, error) {
	var task models.Task
	err := DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	return task, err
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

	userID := middleware.UserID(r)
	task, err := getTask(r, id, userID)
	if err != nil {
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
	if task.Status == "Completed" && task.CompletedAt == nil {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	} else if task.Status != "Completed" {
		task.CompletedAt = nil
	}

	err = DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if task.Status == "Completed" && (!wasCompleted || !hadCompletionTimestamp) {
			return insertCompletion(tx, task)
		}
		if task.Status == "Completed" && actualMinutesProvided {
			var completion models.StudyCompletion
			err := tx.Where(
				"user_id = ? AND task_id = ? AND completed_at = ?",
				task.UserID, task.ID, *task.CompletedAt,
			).First(&completion).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return insertCompletion(tx, task)
			}
			if err != nil {
				return err
			}
			return tx.Model(&completion).Update("actual_study_minutes", task.ActualStudyMinutes).Error
		}
		return nil
	})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not update task")
		return
	}
	JSONResponse(w, http.StatusOK, task)
}

func insertCompletion(tx *gorm.DB, task models.Task) error {
	return tx.Create(&models.StudyCompletion{
		UserID:             task.UserID,
		TaskID:             task.ID,
		ActualStudyMinutes: task.ActualStudyMinutes,
		CompletedAt:        *task.CompletedAt,
	}).Error
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
	result := DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", id, middleware.UserID(r)).Delete(&models.Task{})
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
	userID := middleware.UserID(r)
	task, err := getTask(r, id, userID)
	if err != nil {
		writeTaskLookupError(w, err)
		return
	}
	wasCompleted := task.Status == "Completed" && task.CompletedAt != nil
	if !wasCompleted {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	task.Status = "Completed"

	if err := DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		if !wasCompleted {
			return insertCompletion(tx, task)
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
