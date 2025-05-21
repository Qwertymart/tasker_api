package service

import (
	"errors"
	"gorm.io/gorm"
	"tasker_api/task/internal/model"
	"time"
)

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) GetTaskByUser(userId uint) ([]model.Task, error) {
	var tasks []model.Task
	if err := s.db.Where("user_id = ?", userId).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) CreateTask(task *model.Task) error {
	if task.Title == "" {
		return errors.New("empty title")
	}

	if task.Deadline != nil && time.Now().After(*task.Deadline) {
		return errors.New("deadline cannot be in the past")
	}

	task.IsReady = false
	return s.db.Create(task).Error
}

func (s *TaskService) DeleteTask(taskID, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", taskID, userID).Delete(&model.Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found or does not belong to user")
	}
	return nil
}

func (s *TaskService) UpdateTask(taskID, userID uint, newTitle, newDescription string, newDeadline *time.Time) error {
	var task model.Task

	if err := s.db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		return err
	}

	if newTitle != "" {
		task.Title = newTitle
	}
	if newDescription != "" {
		task.Description = newDescription
	}
	if newDeadline != nil && time.Now().After(*newDeadline) {
		task.Deadline = newDeadline
	}

	return s.db.Save(&task).Error
}

func (s *TaskService) UpdateStateTask(taskID, userID uint, isReady bool) error {
	var task model.Task

	if err := s.db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		return err
	}

	task.IsReady = isReady

	return s.db.Save(&task).Error
}
