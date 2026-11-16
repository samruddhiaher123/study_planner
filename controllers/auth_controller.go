package controllers

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"study-planner/dto"
	"study-planner/services"
	"study-planner/utils"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register creates an account.
// @Summary Register a user
// @Description Creates an account with a name, email address, and password.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest
	if !utils.DecodeJSON(w, r, &input) {
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Name == "" || !validEmailAddress(input.Email) || len([]rune(input.Name)) > 200 ||
		len([]rune(input.Email)) > 320 || len(input.Password) < 6 || len(input.Password) > 72 {
		utils.JSONError(w, http.StatusBadRequest, "Name and a valid email are required; password must be 6 to 72 characters")
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
	utils.JSONResponse(w, http.StatusCreated, dto.RegisterResponse{Message: "Registration successful"})
}

// Login authenticates a user and returns a JWT.
// @Summary Log in
// @Description Authenticates a user and returns a JWT for protected endpoints.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest
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
	utils.JSONResponse(w, http.StatusOK, dto.AuthResponse{
		Message: "Login successful",
		Token:   token,
	})
}

func validEmailAddress(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
