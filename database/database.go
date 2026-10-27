package database

import (
	"fmt"

	"study-planner/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting database connection: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("database is not reachable: %w", err)
	}
	return db, nil
}

func BackfillLegacyData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Task{}).
			Where("status IS NULL OR status = ? OR status = ?", "", "Pending").
			Update("status", "Not Started").Error; err != nil {
			return fmt.Errorf("task status update failed: %w", err)
		}
		if err := tx.Model(&models.Task{}).
			Where("priority IS NULL OR priority = ?", "").
			Update("priority", "Medium").Error; err != nil {
			return fmt.Errorf("task priority update failed: %w", err)
		}

		var tasks []models.Task
		err := tx.Where(
			"status = ? AND completed_at IS NOT NULL AND NOT EXISTS "+
				"(SELECT 1 FROM study_completions "+
				"WHERE study_completions.task_id = tasks.id "+
				"AND study_completions.user_id = tasks.user_id "+
				"AND study_completions.completed_at = tasks.completed_at)",
			"Completed",
		).Find(&tasks).Error
		if err != nil {
			return fmt.Errorf("finding missing task history: %w", err)
		}

		for _, task := range tasks {
			if task.CompletedAt == nil {
				continue
			}
			if err := tx.Create(&models.StudyCompletion{
				UserID:      task.UserID,
				TaskID:      task.ID,
				CompletedAt: *task.CompletedAt,
			}).Error; err != nil {
				return fmt.Errorf("saving task history failed: %w", err)
			}
		}
		return nil
	})
}
