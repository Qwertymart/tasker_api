package service

import (
	"errors"
	"gorm.io/gorm"
	"tasker_api/user/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user *model.User

	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(user *model.User) error {
	if user.Password == "" {
		return errors.New("empty password")
	}

	if user.Username == "" {
		return errors.New("empty username")
	}
	var search model.User
	if err := s.db.Where("username = ?", user.Username).First(&search).Error; err == nil {
		return errors.New("username taken")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("database error")
	}

	return s.db.Create(&user).Error
}

func (s *UserService) DeleteUser(user *model.User) error {
	result := s.db.Where("id = ?", user.ID).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user + not found")
	}
	return nil
}
