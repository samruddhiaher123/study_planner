package main

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
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Could not load .env file: %v", err)
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is empty; set it in .env before starting the server")
	}

	db, err := config.Connect()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Task{}); err != nil {
		log.Fatal(err)
	}
	handlers.DB = db

	router := mux.NewRouter()
	router.Use(cors)
	router.HandleFunc("/api/register", handlers.Register).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.Auth)
	protected.HandleFunc("/tasks", handlers.GetTasks).Methods("GET")
	protected.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
	protected.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	protected.HandleFunc("/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", handlers.DeleteTask).Methods("DELETE")
	protected.HandleFunc("/tasks/{id}/complete", handlers.CompleteTask).Methods("PUT")
	protected.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
	})
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
