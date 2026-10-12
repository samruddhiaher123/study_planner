package main

// @title Study Planner API
// @version 1.0
// @description API for managing study sessions and tasks
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"log"
	"net/http"
	"os"

	"study-planner/config"
	"study-planner/handlers"
	"study-planner/middleware"
	"study-planner/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "study-planner/docs"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Could not load .env file: %v", err)
	}

	// Check JWT secret
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is empty; set it in .env before starting the server")
	}

	// Connect to database
	db, err := config.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Create database tables if needed
	if err := db.AutoMigrate(&models.User{}, &models.Task{}, &models.StudyCompletion{}); err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&models.Task{}).
		Where("status IS NULL OR status = ? OR status = ?", "", "Pending").
		Update("status", "Not Started").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Model(&models.Task{}).
		Where("priority IS NULL OR priority = ?", "").
		Update("priority", "Medium").Error; err != nil {
		log.Fatal(err)
	}
	var legacyCompletions []models.Task
	if err := db.Where(
		"status = ? AND completed_at IS NOT NULL AND NOT EXISTS "+
			"(SELECT 1 FROM study_completions WHERE study_completions.task_id = tasks.id "+
			"AND study_completions.user_id = tasks.user_id AND study_completions.completed_at = tasks.completed_at)",
		"Completed",
	).Find(&legacyCompletions).Error; err != nil {
		log.Fatal(err)
	}
	for _, task := range legacyCompletions {
		if err := db.Create(&models.StudyCompletion{
			UserID:             task.UserID,
			TaskID:             task.ID,
			ActualStudyMinutes: task.ActualStudyMinutes,
			CompletedAt:        *task.CompletedAt,
		}).Error; err != nil {
			log.Fatal(err)
		}
	}

	handlers.DB = db

	// Initialize Gorilla Mux router
	router := mux.NewRouter()

	// Swagger UI route
	router.PathPrefix("/swagger/").Handler(
		httpSwagger.WrapHandler,
	)

	// CORS middleware
	router.Use(cors)

	// Public authentication routes
	router.HandleFunc("/api/register", handlers.Register).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")

	// Protected API routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.Auth)

	protected.HandleFunc("/tasks", handlers.GetTasks).Methods("GET")
	protected.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
	protected.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	protected.HandleFunc("/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", handlers.DeleteTask).Methods("DELETE")
	protected.HandleFunc("/tasks/{id}/complete", handlers.CompleteTask).Methods("PUT")
	protected.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")

	// Frontend routes
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
	})
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	// Start server
	log.Println("Swagger UI: http://localhost:8080/swagger/index.html")
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// CORS middleware
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "X-Server-Date")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
