package handler

import (
	"auth/pkg/auth_user_pb"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"time"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtSecret = []byte(secret)
}

type AuthHandler struct {
	authClient auth_user_pb.AuthServiceClient
}

func NewAuthHandler(authClient auth_user_pb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		RepeatPassword string `json:"repeat_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	if input.RepeatPassword != input.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	grpcReq := &auth_user_pb.RegisterRequest{
		Username: input.Username,
		Password: input.Password,
	}

	res, err := h.authClient.Register(c, grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !res.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully,",
		"id": res.Id})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	grpcReq := &auth_user_pb.LoginRequest{
		Username: input.Username,
		Password: input.Password,
	}

	res, err := h.authClient.Login(c, grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error: " + err.Error()})
		return
	}

	if !res.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": res.Error})
		return
	}

	token, err := generateJWT(uint(res.Id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Login successful",
		"id":      res.Id,
	})

}

func generateJWT(id uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Life of token 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
