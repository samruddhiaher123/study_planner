package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"study-planner/models"
	"study-planner/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
func Register(w http.ResponseWriter, r *http.Request) {
	var input registerRequest
	if !DecodeJSON(w, r, &input) {
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Name == "" || input.Email == "" || len(input.Password) < 6 || len(input.Password) > 72 {
		JSONError(w, http.StatusBadRequest, "Name and email are required; password must be 6 to 72 characters")
		return
	}

	var existing models.User
	if err := DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		JSONError(w, http.StatusBadRequest, "Email already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		JSONError(w, http.StatusInternalServerError, "Could not check email")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not save user")
		return
	}

	user := models.User{Name: input.Name, Email: input.Email, Password: string(hashedPassword)}
	if err := DB.Create(&user).Error; err != nil {
		JSONError(w, http.StatusInternalServerError, "Could not save user")
		return
	}

	JSONResponse(w, http.StatusCreated, registerResponse{Message: "Registration successful"})
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
func Login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if !DecodeJSON(w, r, &input) {
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	var user models.User
	if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			JSONError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			JSONError(w, http.StatusInternalServerError, "Could not look up user")
		}
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		JSONError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		log.Println("JWT generation error:", err)
		JSONError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}
	JSONResponse(w, http.StatusOK, loginResponse{
		Message: "Login successful",
		Token:   token,
	})
}
