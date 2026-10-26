package services

import (
	"context"
	"errors"
	"testing"

	"study-planner/models"
	"study-planner/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepositoryStub struct {
	user models.User
}

func (r *userRepositoryStub) GetByEmail(_ context.Context, email string) (models.User, error) {
	if r.user.Email != email {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return r.user, nil
}

func (r *userRepositoryStub) Create(_ context.Context, user *models.User) error {
	r.user = *user
	r.user.ID = 42
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret")
	repository := &userRepositoryStub{}
	service := NewAuthService(repository)

	if err := service.Register(context.Background(), " Alex ", "ALEX@example.com ", "study123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.user.Name != "Alex" || repository.user.Email != "alex@example.com" {
		t.Errorf("registered user = %+v, expected normalized name and email", repository.user)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.user.Password), []byte("study123")); err != nil {
		t.Errorf("stored password is not a valid bcrypt hash: %v", err)
	}
	if repository.user.Password == "study123" {
		t.Error("password was stored in plaintext")
	}

	token, err := service.Login(context.Background(), " ALEX@example.com ", "study123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	userID, err := utils.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if userID != 42 {
		t.Errorf("token user ID = %d, want 42", userID)
	}

	_, err = service.Login(context.Background(), "alex@example.com", "incorrect")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}
