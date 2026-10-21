package config

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"study-planner/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5432")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	sslMode := envOrDefault("DB_SSLMODE", "disable")

	// Check database details.
	if user == "" {
		return nil, fmt.Errorf("DB_USER is empty")
	}
	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is empty")
	}
	if name == "" {
		return nil, fmt.Errorf("DB_NAME is empty")
	}

	// Create the PostgreSQL connection URL.
	query := url.Values{}
	query.Set("sslmode", sslMode)

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + name,
		RawQuery: query.Encode(),
	}

	// Connect to the database.
	db, err := gorm.Open(postgres.Open(dbURL.String()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Check whether the database is working.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("database is not reachable: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	// Create or update the required tables.
	err := db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.StudyCompletion{},
	)
	if err != nil {
		return fmt.Errorf("table creation failed: %w", err)
	}

	// Update old task data.
	return db.Transaction(func(tx *gorm.DB) error {

		// Change empty or Pending status to Not Started.
		err := tx.Model(&models.Task{}).
			Where("status IS NULL OR status = ? OR status = ?", "", "Pending").
			Update("status", "Not Started").Error
		if err != nil {
			return fmt.Errorf("task status update failed: %w", err)
		}

		// Set Medium priority when priority is empty.
		err = tx.Model(&models.Task{}).
			Where("priority IS NULL OR priority = ?", "").
			Update("priority", "Medium").Error
		if err != nil {
			return fmt.Errorf("task priority update failed: %w", err)
		}

		// Find completed tasks whose history is missing.
		var tasks []models.Task

		err = tx.Where(
			"status = ? AND completed_at IS NOT NULL AND NOT EXISTS "+
				"(SELECT 1 FROM study_completions "+
				"WHERE study_completions.task_id = tasks.id "+
				"AND study_completions.user_id = tasks.user_id "+
				"AND study_completions.completed_at = tasks.completed_at)",
			"Completed",
		).Find(&tasks).Error

		if err != nil {
			return fmt.Errorf("could not find old task history: %w", err)
		}

		// Save the missing task history.
		for _, task := range tasks {
			if task.CompletedAt == nil {
				continue
			}

			err = tx.Create(&models.StudyCompletion{
				UserID:             task.UserID,
				TaskID:             task.ID,
				ActualStudyMinutes: task.ActualStudyMinutes,
				CompletedAt:        *task.CompletedAt,
			}).Error

			if err != nil {
				return fmt.Errorf("could not save task history: %w", err)
			}
		}

		return nil
	})
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)

	if value == "" {
		return fallback
	}

	return value
}
