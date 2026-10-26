package repositories

import (
	"context"
	"strings"
	"time"

	"study-planner/models"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	List(ctx context.Context, userID uint, filter models.TaskFilter) ([]models.Task, error)
	GetByID(ctx context.Context, id, userID uint) (models.Task, error)
	Update(ctx context.Context, task *models.Task, addCompletion bool) error
	Delete(ctx context.Context, id, userID uint) (bool, error)
	ListForDashboard(ctx context.Context, userID uint) ([]models.Task, error)
	CompletionHistory(ctx context.Context, userID uint) ([]models.StudyCompletion, error)
}

type GormTaskRepository struct {
	db *gorm.DB
}

func NewGormTaskRepository(db *gorm.DB) *GormTaskRepository {
	return &GormTaskRepository{db: db}
}

func (r *GormTaskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		if task.CompletedAt == nil {
			return nil
		}
		return tx.Create(&models.StudyCompletion{
			UserID:      task.UserID,
			TaskID:      task.ID,
			CompletedAt: *task.CompletedAt,
		}).Error
	})
}

func (r *GormTaskRepository) List(ctx context.Context, userID uint, filter models.TaskFilter) ([]models.Task, error) {
	conditions := []string{"user_id = ?"}
	args := []any{userID}
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, condition)
	}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
		conditions = append(conditions, "(title LIKE ? OR description LIKE ?)")
	}
	if filter.Subject != "" {
		addCondition("subject = ?", filter.Subject)
	}
	if filter.Status != "" {
		addCondition("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		addCondition("priority = ?", filter.Priority)
	}
	if filter.Date != "" {
		addCondition("date = ?", filter.Date)
	}
	if filter.From != "" {
		addCondition("date >= ?", filter.From)
	}
	if filter.To != "" {
		addCondition("date <= ?", filter.To)
	}
	switch filter.View {
	case "", "all":
	case "completed":
		conditions = append(conditions, "status = 'Completed'")
	case "pending":
		conditions = append(conditions, "status <> 'Completed'")
	case "overdue":
		addCondition("date < ? AND status <> 'Completed'", time.Now().Format("2006-01-02"))
	case "upcoming":
		addCondition("date >= ? AND status <> 'Completed'", time.Now().Format("2006-01-02"))
	}

	query := `
		SELECT id, user_id, title, subject, description, date, start_time, end_time,
		       status, priority, completed_at, created_at
		FROM tasks
		WHERE ` + strings.Join(conditions, " AND ")
	today := time.Now().Format("2006-01-02")
	switch filter.Sort {
	case "", "default", "dueDate":
		query += " ORDER BY CASE WHEN status = 'Completed' THEN 1 ELSE 0 END,"
		query += " CASE WHEN status <> 'Completed' AND date < '" + today + "' THEN 0 ELSE 1 END,"
		query += " CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END,"
		query += " date, start_time, id"
	case "priority":
		query += " ORDER BY CASE WHEN status = 'Completed' THEN 1 ELSE 0 END,"
		query += " CASE priority WHEN 'High' THEN 0 WHEN 'Medium' THEN 1 ELSE 2 END,"
		query += " date, start_time, id"
	case "newest":
		query += " ORDER BY created_at DESC, id DESC"
	case "oldest":
		query += " ORDER BY created_at ASC, id ASC"
	case "status":
		query += " ORDER BY status, date, start_time, id"
	}

	tasks := make([]models.Task, 0)
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&tasks).Error
	return tasks, err
}

func (r *GormTaskRepository) GetByID(ctx context.Context, id, userID uint) (models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	return task, err
}

func (r *GormTaskRepository) Update(ctx context.Context, task *models.Task, addCompletion bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(task).Error; err != nil {
			return err
		}
		if !addCompletion {
			return nil
		}
		return tx.Create(&models.StudyCompletion{
			UserID:      task.UserID,
			TaskID:      task.ID,
			CompletedAt: *task.CompletedAt,
		}).Error
	})
}

func (r *GormTaskRepository) Delete(ctx context.Context, id, userID uint) (bool, error) {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{})
	return result.RowsAffected > 0, result.Error
}

func (r *GormTaskRepository) ListForDashboard(ctx context.Context, userID uint) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *GormTaskRepository) CompletionHistory(ctx context.Context, userID uint) ([]models.StudyCompletion, error) {
	var completions []models.StudyCompletion
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&completions).Error; err != nil {
		return nil, err
	}
	return completions, nil
}
