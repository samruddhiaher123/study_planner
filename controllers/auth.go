package controllers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"study-planner/services"
	"study-planner/utils"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	Message string `json:"message"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

// Register creates an account.
// @Summary Register a user
// @Description Creates an account with a name, email address, and password.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration details"
// @Success 201 {object} registerResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input registerRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Name == "" || input.Email == "" || len([]rune(input.Name)) > 200 ||
		len([]rune(input.Email)) > 320 || len(input.Password) < 6 || len(input.Password) > 72 {
		utils.JSONError(w, http.StatusBadRequest, "Name and email are required; password must be 6 to 72 characters")
		return
	}

	if err := h.service.Register(r.Context(), input.Name, input.Email, input.Password); err != nil {
		if errors.Is(err, services.ErrEmailExists) {
			utils.JSONError(w, http.StatusBadRequest, "Email already exists")
			return
		}
		if errors.Is(err, services.ErrEmailLookup) {
			utils.JSONError(w, http.StatusInternalServerError, "Could not check email")
			return
		}
		utils.JSONError(w, http.StatusInternalServerError, "Could not save user")
		return
	}
	utils.JSONResponse(w, http.StatusCreated, registerResponse{Message: "Registration successful"})
}

// Login authenticates a user and returns a JWT.
// @Summary Log in
// @Description Authenticates a user and returns a JWT for protected endpoints.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} loginResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	token, err := h.service.Login(r.Context(), input.Email, input.Password)
	if errors.Is(err, services.ErrInvalidCredentials) {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if errors.Is(err, services.ErrTokenGeneration) {
		log.Println("JWT generation error:", err)
		utils.JSONError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Could not look up user")
		return
	}
	utils.JSONResponse(w, http.StatusOK, loginResponse{
		Message: "Login successful",
		Token:   token,
	})
}
