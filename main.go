package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"`
	UserID      uint       `json:"user_id"`
	CreatedAt   *time.Time `json:"created_at"`
}

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var Users []User
var Tasks []Task

func home(c *gin.Context) {
	c.JSON(200, "")
}

func login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, u := range Users {
		if u.Username == user.Username && u.Password == user.Password {
			c.JSON(http.StatusOK, gin.H{
				"message": "Login successful",
				"user_id": u.ID,
			})
			return
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid username or password",
	})
}

func register(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, u := range Users {
		if u.Username == user.Username {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username taken",
			})
		}
		return
	}

	user.ID = uint(len(Users) + 1)
	Users = append(Users, user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user_id": user.ID,
	})

}

func getTasks(c *gin.Context) {
	userIDStr := c.Query("user_id")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	userTasks := make([]Task, 0)
	for _, t := range Tasks {
		if uint64(t.UserID) == userIDUint {
			userTasks = append(userTasks, t)
		}
	}

	c.JSON(http.StatusOK, userTasks)
}

func addTask(c *gin.Context) {
	var newTask Task

	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if newTask.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty title",
		})
		return
	}

	if newTask.CreatedAt == nil || newTask.Deadline == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "created_at and deadline must be provided",
		})
		return
	}

	if newTask.CreatedAt.After(*newTask.Deadline) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid deadline: created_at is after deadline",
		})
		return
	}
	newTask.ID = uint(len(Tasks) + 1)
	Tasks = append(Tasks, newTask)
	c.JSON(http.StatusOK, gin.H{
		"task": newTask,
	})
}

func main() {
	r := gin.Default()

	r.GET("/", home)

	r.POST("/login", login)

	r.POST("/register", register)

	r.GET("/tasks", getTasks)

	r.POST("/tasks", addTask)

	r.Run(":8080")

}
