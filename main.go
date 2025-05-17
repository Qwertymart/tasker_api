package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Task struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"`
	UserID      uint       `json:"user_id"`
	CreatedAt   *time.Time `json:"created_at"`
}

var Users []User
var Tasks []Task

var jwtSecret = []byte("secret_key")

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

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
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
			token, err := createToken(u.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate token",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Login successful",
				"user_id": u.ID,
				"token":   token,
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID not found in context"})
		return
	}

	userTasks := make([]Task, 0)
	for _, t := range Tasks {
		if t.UserID == userID.(uint) {
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

	userID, _ := c.Get("userID")
	newTask.ID = uint(len(Tasks) + 1)
	newTask.UserID = userID.(uint)
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

	auth := r.Group("/")
	auth.Use(authMiddleware)
	{
		auth.GET("/tasks", getTasks)
		auth.POST("/tasks", addTask)
	}

	err := r.Run(":8080")
	if err != nil {
		return
	}

}
