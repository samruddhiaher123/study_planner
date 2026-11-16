package routes

import (
	"net/http"

	"study-planner/controllers"
	"study-planner/middleware"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "study-planner/docs"
)

type Controllers struct {
	Auth      *controllers.AuthHandler
	Tasks     *controllers.TaskHandler
	Dashboard *controllers.DashboardHandler
}

func New(api Controllers, staticDir string) http.Handler {
	router := mux.NewRouter()
	router.Use(middleware.CORS)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.HandleFunc("/api/register", api.Auth.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/login", api.Auth.Login).Methods(http.MethodPost)

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.Auth)
	protected.HandleFunc("/tasks", api.Tasks.GetTasks).Methods(http.MethodGet)
	protected.HandleFunc("/tasks", api.Tasks.CreateTask).Methods(http.MethodPost)
	protected.HandleFunc("/tasks/{id}", api.Tasks.GetTask).Methods(http.MethodGet)
	protected.HandleFunc("/tasks/{id}", api.Tasks.UpdateTask).Methods(http.MethodPut)
	protected.HandleFunc("/tasks/{id}", api.Tasks.DeleteTask).Methods(http.MethodDelete)
	protected.HandleFunc("/tasks/{id}/complete", api.Tasks.CompleteTask).Methods(http.MethodPut)
	protected.HandleFunc("/dashboard", api.Dashboard.Dashboard).Methods(http.MethodGet)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusSeeOther)
	})
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))
	return router
}
