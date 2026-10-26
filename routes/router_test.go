package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"study-planner/controllers"
)

func TestRoutesProtectAPIAndServeFrontend(t *testing.T) {
	router := New(Controllers{
		Auth:      controllers.NewAuthHandler(nil),
		Tasks:     controllers.NewTaskHandler(nil),
		Dashboard: controllers.NewDashboardHandler(nil),
	}, "../frontend")

	protectedRequest := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	protectedResponse := httptest.NewRecorder()
	router.ServeHTTP(protectedResponse, protectedRequest)
	if protectedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("protected route status = %d, want %d", protectedResponse.Code, http.StatusUnauthorized)
	}
	if got := protectedResponse.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS origin header = %q, want *", got)
	}

	rootRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	rootResponse := httptest.NewRecorder()
	router.ServeHTTP(rootResponse, rootRequest)
	if rootResponse.Code != http.StatusSeeOther {
		t.Fatalf("root route status = %d, want %d", rootResponse.Code, http.StatusSeeOther)
	}
	if got := rootResponse.Header().Get("Location"); got != "/login.html" {
		t.Errorf("root redirect = %q, want /login.html", got)
	}

	staticRequest := httptest.NewRequest(http.MethodGet, "/login.html", nil)
	staticResponse := httptest.NewRecorder()
	router.ServeHTTP(staticResponse, staticRequest)
	if staticResponse.Code != http.StatusOK {
		t.Fatalf("static route status = %d, want %d", staticResponse.Code, http.StatusOK)
	}
}
