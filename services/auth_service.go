package services

import (
	"context"
	"errors"
	"strings"

	"study-planner/models"
	"study-planner/repositories"
	"study-planner/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrEmailLookup        = errors.New("could not check email")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenGeneration    = errors.New("could not generate token")
)

type AuthService struct {
	users repositories.UserRepository
}

func NewAuthService(users repositories.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	var err error
	if _, err = s.users.GetByEmail(ctx, email); err == nil {
		return ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Join(ErrEmailLookup, err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{Name: name, Email: email, Password: string(hashedPassword)}
	if err := s.users.Create(ctx, &user); err != nil {
		if _, lookupErr := s.users.GetByEmail(ctx, email); lookupErr == nil {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", errors.Join(ErrTokenGeneration, err)
	}
	return token, nil
}
