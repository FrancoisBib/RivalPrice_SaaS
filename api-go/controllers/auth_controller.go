package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/middleware"
	"github.com/rivalprice/api-go/services"
)

type AuthController struct {
	userService *services.UserService
	jwtSecret   string
}

func NewAuthController(userService *services.UserService, jwtSecret string) *AuthController {
	return &AuthController{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Login - POST /auth/login
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement password verification (currently not stored)
	user, err := c.userService.GetUserByEmail(req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token (valid for 24 hours)
	token, err := middleware.GenerateToken(user.ID, user.Email, c.jwtSecret, 24)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

// Register - POST /auth/register
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	_, err := c.userService.GetUserByEmail(req.Email)
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create user
	user, err := c.userService.CreateUser(req.Name, req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, c.jwtSecret, 24)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

// Me - GET /auth/me (protected)
func (c *AuthController) Me(ctx *gin.Context) {
	userID, exists := middleware.GetUserID(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := c.userService.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}
