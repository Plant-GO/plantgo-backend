package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"

    "plantgo-backend/internal/dto"
    "plantgo-backend/internal/modules/auth/infrastructure"
)

type AuthService struct {
	userRepo *infrastructure.UserRepository
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		userRepo: infrastructure.NewUserRepository(db),
	}
}

// GuestLoginHandler godoc
// @Summary      Guest login
// @Description  Authenticates or creates a guest user using Android ID and username
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.GuestLoginRequest true "Guest login credentials"
// @Success      200 {object} dto.AuthResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /auth/guest/login [post]
func (s *AuthService) GuestLoginHandler(c *gin.Context) {
	var req dto.GuestLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user, exists := s.userRepo.UserExists("", req.AndroidID)
	if exists {
		if user.Username != req.Username {
			user.Username = req.Username
			user.UpdatedAt = time.Now().UTC()
			if err := s.userRepo.UpdateUser(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
				return
			}
		}
	} else {
		user = &infrastructure.User{
			AndroidID:  &req.AndroidID,
			Username:   req.Username,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}

		if err := s.userRepo.CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest user"})
			return
		}
	}

	jwtToken, err := generateJWT(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": user})
}

// GoogleLoginHandler godoc
// @Summary      Initiate Google OAuth login
// @Description  Redirects the user to Google's OAuth2 authorization page
// @Tags         Auth
// @Produce      plain
// @Success      307 {string} string "Redirects to Google OAuth2 page"
// @Router       /auth/google/login [get]
func (s *AuthService) GoogleLoginHandler(c *gin.Context) {
	url := GoogleOAuthConfig.AuthCodeURL("state-token")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallbackHandler godoc
// @Summary      Handle Google OAuth callback
// @Description  Processes the OAuth2 callback from Google and returns a JWT token
// @Tags         Auth
// @Produce      json
// @Param        code query string true "Authorization code from Google"
// @Param        state query string true "State token"
// @Success      200 {object} dto.AuthResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /auth/google/callback [get]
func (s *AuthService) GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")

	token, err := GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token exchange failed"})
		return
	}

	client := GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var userInfo map[string]interface{}
	json.Unmarshal(body, &userInfo)

	googleID := userInfo["id"].(string)
	email := userInfo["email"].(string)
	username := userInfo["name"].(string)

	user := &infrastructure.User{
		GoogleID:  &googleID,
		Email:     email,
		Username:  username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Create or update user in database
	savedUser, err := s.userRepo.CreateOrUpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user", "details": err.Error()})
		return
	}

	jwtToken, err := generateJWT(*savedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": savedUser})
}

// RegisterHandler godoc
// @Summary      Register a new user
// @Description  Creates a new user with username, email, and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "infrastructure.User registration info"
// @Success      201 {object} dto.AuthResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /auth/register [post]
func (s *AuthService) RegisterHandler(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exists := s.userRepo.UserExists(req.Email, ""); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "infrastructure.User with this email already exists"})
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &infrastructure.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	jwtToken, err := generateJWT(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": jwtToken, "user": user})
}

// LoginHandler godoc
// @Summary      Login
// @Description  Login with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.AuthResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /auth/login [post]
func (s *AuthService) LoginHandler(c *gin.Context) {
    // Login handler for email/password authentication

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password 
	if !verifyPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	jwtToken, err := generateJWT(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": user})
}

// GetProfileHandler godoc
// @Summary      Get user profile
// @Description  Returns the authenticated user's profile
// @Tags         Auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} infrastructure.User
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /profile [get]
func (s *AuthService) GetProfileHandler(c *gin.Context) {
    // Get user profile
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := s.userRepo.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func generateJWT(user infrastructure.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"email":    user.Email,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// hashPassword hashes a plain text password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword compares a plain text password with a bcrypt hash
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}