package handler

import (
	"auth/pkg/auth_user_pb"
	"auth/pkg/google_oauth"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := google_oauth.GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found in callback"})
		return
	}

	token, err := google_oauth.GoogleOAuthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	client := google_oauth.GoogleOAuthConfig.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user info response"})
		return
	}

	res, err := h.authClient.LoginWithGoogle(c, &auth_user_pb.GoogleLoginRequest{
		GoogleId: userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
	})
	if err != nil || !res.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth failed"})
		return
	}

	jwtToken, err := generateJWT(uint(res.Id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   jwtToken,
		"message": "Google login successful",
		"id":      res.Id,
	})
}
