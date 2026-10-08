package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"study-planner/utils"
)

func TestAuthAddsUserIDToRequest(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	const wantUserID = uint(42)

	token, err := utils.GenerateToken(wantUserID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatUint(uint64(UserID(r)), 10)))
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Auth() status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "42" {
		t.Fatalf("UserID() = %q, want %q", response.Body.String(), "42")
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("protected handler ran without a token")
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("Auth() status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
