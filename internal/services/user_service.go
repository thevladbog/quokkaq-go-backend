package services

import (
	"errors"
	"quokkaq-go-backend/internal/models"
	"quokkaq-go-backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *models.User) error
	GetAllUsers() ([]models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	AssignUnit(userID, unitID string, permissions []string) error
	RemoveUnit(userID, unitID string) error
	AssignRole(userID, roleID string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(user *models.User) error {
	// Check if email exists
	if user.Email != nil {
		existing, _ := s.repo.FindByEmail(*user.Email)
		if existing != nil {
			return errors.New("email already in use")
		}
	}

	// Hash password if present
	if user.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashedStr := string(hashed)
		user.Password = &hashedStr
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	return s.repo.Create(user)
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) UpdateUser(user *models.User) error {
	if user.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashedStr := string(hashed)
		user.Password = &hashedStr
	}
	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

func (s *userService) AssignUnit(userID, unitID string, permissions []string) error {
	return s.repo.AssignUnit(userID, unitID, permissions)
}

func (s *userService) AssignRole(userID, roleID string) error {
	return s.repo.AssignRole(userID, roleID)
}

func (s *userService) RemoveUnit(userID, unitID string) error {
	return s.repo.RemoveUnit(userID, unitID)
}
