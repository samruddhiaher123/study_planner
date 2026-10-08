package handlers

import (
	"errors"
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
	Title       string `json:"title"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Date        string `json:"date"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Status      string `json:"status"`
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
		input.Status = "Pending"
	}
	if input.Status != "Pending" && input.Status != "Completed" {
		return "Status must be Pending or Completed"
	}
	return ""
}

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
	}
	if err := DB.Create(&task).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not create task")
		return
	}
	JSONResponse(w, http.StatusCreated, task)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []models.Task
	if err := DB.Where("user_id = ?", middleware.UserID(r)).Order("date ASC, start_time ASC").Find(&tasks).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not get tasks")
		return
	}
	JSONResponse(w, http.StatusOK, tasks)
}

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

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	var input taskRequest
	if !DecodeJSON(w, r, &input) {
		return
	}
	if message := validateTask(&input); message != "" {
		JSONError(w, http.StatusBadRequest, message)
		return
	}

	var task models.Task
	if err := DB.Where("id = ? AND user_id = ?", id, middleware.UserID(r)).First(&task).Error; err != nil {
		writeTaskLookupError(w, err)
		return
	}
	task.Title = input.Title
	task.Subject = input.Subject
	task.Description = input.Description
	task.Date = input.Date
	task.StartTime = input.StartTime
	task.EndTime = input.EndTime
	task.Status = input.Status

	if err := DB.Save(&task).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not update task")
		return
	}
	JSONResponse(w, http.StatusOK, task)
}

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
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Task deleted"})
}

func CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := taskID(w, r)
	if !ok {
		return
	}
	result := DB.Model(&models.Task{}).
		Where("id = ? AND user_id = ?", id, middleware.UserID(r)).
		Update("status", "Completed")
	if result.Error != nil {
		JSONError(w, http.StatusInternalServerError, "Could not complete task")
		return
	}
	if result.RowsAffected == 0 {
		JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "Task completed"})
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
