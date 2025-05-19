package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Tasks    []Task `gorm:"foreignKey:UserID"`
}

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtSecret = []byte(secret)
}

var DB *gorm.DB

func connectDB() {
	dsn := "host=localhost user=postgres password=5432 dbname=task_manager port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&User{}, &Task{})
	if err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	DB = db
}

func createToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), //Lives 24 hours
	}
	//create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	return token.SignedString(jwtSecret)
}

func authMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header required",
		})
		c.Abort()
		return
	}

	const prefix = "Bearer "

	if len(prefix) >= len(authHeader) || prefix != authHeader[:len(prefix)] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization header must be bearer",
		})
		c.Abort()
		return
	}

	tokenString := authHeader[len(prefix):]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // Checking the algorithm
			return nil, jwt.ErrInvalidKey
		}
		return jwtSecret, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to parse token: " + err.Error(),
		})
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
		c.Abort()
		return
	}
	userID := uint(userIDFloat)

	c.Set("userID", userID)

	c.Next()
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}

func login(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var user User
	if err := DB.Where("username = ? AND password = ?", input.Username, input.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := createToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user_id": user.ID,
		"token":   token,
	})
}

func register(c *gin.Context) {
	var input User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user User

	if err := DB.Where("username = ?", input.Username).First(&user).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username taken",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	user = User{
		Username: input.Username,
		Password: input.Password,
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user_id": user.ID,
	})

}

func getTasks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID not found in context"})
		return
	}

	var tasks []Task
	if err := DB.Where("user_id = ?", userID.(uint)).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func addTask(c *gin.Context) {
	var newTask Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newTask.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty title"})
		return
	}

	if newTask.Deadline != nil && time.Now().After(*newTask.Deadline) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deadline cannot be in the past"})
		return
	}

	userID, _ := c.Get("userID")
	newTask.UserID = userID.(uint)

	if err := DB.Create(&newTask).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": newTask})
}

func deleteTask(c *gin.Context) {
	var input struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID has invalid type"})
		return
	}

	result := DB.Where("id = ? AND user_id = ?", input.ID, userID).Delete(&Task{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or does not belong to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
		"id":      input.ID,
	})
}

func main() {
	connectDB()
	r := gin.Default()

	r.GET("/", home)
	r.POST("/login", login)
	r.POST("/register", register)

	auth := r.Group("/")
	auth.Use(authMiddleware)
	{
		auth.GET("/tasks", getTasks)
		auth.POST("/tasks", addTask)
		auth.DELETE("/tasks", deleteTask)
	}

	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}

}
