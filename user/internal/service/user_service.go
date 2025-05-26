package service

import (
	"errors"
	"gorm.io/gorm"
	"user/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CheckByID(userID uint) (bool, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New("database error")
	}
	return true, nil
}

func (s *UserService) CreateUser(user *model.User) error {
	if user.Username == "" {
		return errors.New("empty username")
	}
	if user.Password == "" {
		return errors.New("empty password")
	}

	var existing model.User
	if err := s.db.Where("username = ?", user.Username).First(&existing).Error; err == nil {
		return errors.New("username taken")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("database error")
	}

	return s.db.Create(user).Error
}

func (s *UserService) DeleteUser(userID uint) error {
	result := s.db.Delete(&model.User{}, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *UserService) UpdateUser(userID uint, updateUser *model.User) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if updateUser.Username != "" {
		user.Username = updateUser.Username
	}
	if updateUser.Password != "" {
		user.Password = updateUser.Password
	}

	return s.db.Save(&user).Error
}
