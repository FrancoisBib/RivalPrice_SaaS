package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/services"
)

type ProjectController struct {
	projectService *services.ProjectService
}

func NewProjectController(projectService *services.ProjectService) *ProjectController {
	return &ProjectController{projectService: projectService}
}

type CreateProjectRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

// CreateProject - POST /projects
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	var req CreateProjectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := c.projectService.CreateProject(req.UserID, req.Name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": project,
	})
}

// ListProjects - GET /projects
func (c *ProjectController) ListProjects(ctx *gin.Context) {
	// Optional: filter by user_id
	userID := ctx.Query("user_id")
	if userID != "" {
		uid, err := strconv.ParseUint(userID, 10, 32)
		if err == nil {
			projects, err := c.projectService.GetProjectsByUserID(uint(uid))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"projects": projects})
			return
		}
	}

	projects, err := c.projectService.GetAllProjects()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

// GetProject - GET /projects/:id
func (c *ProjectController) GetProject(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := c.projectService.GetProjectByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"project": project})
}
