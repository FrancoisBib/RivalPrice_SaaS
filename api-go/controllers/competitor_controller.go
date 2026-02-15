package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/services"
)

type CompetitorController struct {
	competitorService *services.CompetitorService
}

func NewCompetitorController(competitorService *services.CompetitorService) *CompetitorController {
	return &CompetitorController{competitorService: competitorService}
}

type CreateCompetitorRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	URL       string `json:"url"`
}

// CreateCompetitor - POST /competitors
func (c *CompetitorController) CreateCompetitor(ctx *gin.Context) {
	var req CreateCompetitorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	competitor, err := c.competitorService.CreateCompetitor(req.ProjectID, req.Name, req.URL)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Competitor created successfully",
		"competitor": competitor,
	})
}

// ListCompetitors - GET /competitors
func (c *CompetitorController) ListCompetitors(ctx *gin.Context) {
	// Optional: filter by project_id
	projectID := ctx.Query("project_id")
	if projectID != "" {
		pid, err := strconv.ParseUint(projectID, 10, 32)
		if err == nil {
			competitors, err := c.competitorService.GetCompetitorsByProjectID(uint(pid))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch competitors"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"competitors": competitors})
			return
		}
	}

	competitors, err := c.competitorService.GetAllCompetitors()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch competitors"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"competitors": competitors,
	})
}

// GetCompetitor - GET /competitors/:id
func (c *CompetitorController) GetCompetitor(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid competitor ID"})
		return
	}

	competitor, err := c.competitorService.GetCompetitorByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"competitor": competitor})
}
