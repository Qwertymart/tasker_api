package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"task/internal/model"
	"task/internal/service"
	"task/pkg/userpb"
	"time"
)

type TaskHandler struct {
	s          *service.TaskService
	userClient userpb.UserServiceClient
}

func NewTaskHandler(db *gorm.DB, userClient userpb.UserServiceClient) *TaskHandler {
	return &TaskHandler{
		s:          service.NewTaskService(db),
		userClient: userClient,
	}
}

func (h *TaskHandler) validateUser(c *gin.Context) (uint, bool) {

	// потом убрать
	userIDStr := c.GetHeader("userID")
	userIDUint64, err := strconv.ParseUint(userIDStr, 10, 64)
	userID := uint(userIDUint64)

	c.Set("userID", userID)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID missing"})
		return 0, false
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userID"})
		return 0, false
	}

	resp, err := h.userClient.GetUser(context.Background(), &userpb.GetUserRequest{
		Id: strconv.Itoa(int(userID)),
	})
	if err != nil || !resp.Exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return 0, false
	}

	return userID, true
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := h.validateUser(c)
	if !ok {
		return
	}

	tasks, err := h.s.GetTaskByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) AddTask(c *gin.Context) {
	var newTask model.Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := h.validateUser(c)
	if !ok {
		return
	}

	newTask.UserID = userID

	if err := h.s.CreateTask(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": newTask})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	var input struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := h.validateUser(c)
	if !ok {
		return
	}

	if err := h.s.DeleteTask(input.ID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted successfully", "id": input.ID})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var input struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Deadline    *time.Time `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	userID, ok := h.validateUser(c)
	if !ok {
		return
	}

	if err = h.s.UpdateTask(uint(taskID), userID, input.Title, input.Description, input.Deadline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task updated successfully"})
}

func (h *TaskHandler) UpdateStateTask(c *gin.Context) {
	var input struct {
		IsReady bool `json:"is_ready" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	userID, ok := h.validateUser(c)
	if !ok {
		return
	}

	if err = h.s.UpdateStateTask(uint(taskID), userID, input.IsReady); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task state updated successfully"})
}
