// @title Study Planner API
// @version 1.0
// @description API for managing study sessions and tasks
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"study-planner/config"
	"study-planner/controllers"
	"study-planner/database"
	"study-planner/repositories"
	"study-planner/routes"
	"study-planner/services"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Could not load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is empty; set it in .env before starting the server")
	}
	//db connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := database.BackfillLegacyData(db); err != nil {
		log.Fatal(err)
	}

	userRepository := repositories.NewGormUserRepository(db)
	taskRepository := repositories.NewGormTaskRepository(db)
	authService := services.NewAuthService(userRepository)
	taskService := services.NewTaskService(taskRepository)
	dashboardService := services.NewDashboardService(taskRepository)
	apiControllers := routes.Controllers{
		Auth:      controllers.NewAuthHandler(authService),
		Tasks:     controllers.NewTaskHandler(taskService),
		Dashboard: controllers.NewDashboardHandler(dashboardService),
	}
	router := routes.New(apiControllers, cfg.StaticDir)

	log.Printf("Swagger UI: http://localhost:8080/swagger/index.html")
	log.Printf("Server running on %s", cfg.HTTPAddress)
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
