package controllers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/models"
	"github.com/rivalprice/api-go/services"
	"github.com/rivalprice/api-go/utils"
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
	// Get pagination params
	pagination := utils.GetPaginationParams(ctx)
	
	var competitors []models.Competitor
	var total int64
	var err error
	
	// Optional: filter by project_id
	projectID := ctx.Query("project_id")
	if projectID != "" {
		pid, parseErr := strconv.ParseUint(projectID, 10, 32)
		if parseErr == nil {
			competitors, total, err = c.competitorService.GetCompetitorsByProjectIDPaginated(uint(pid), pagination.Offset, pagination.PageSize)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
			return
		}
	} else {
		competitors, total, err = c.competitorService.GetAllCompetitorsPaginated(pagination.Offset, pagination.PageSize)
	}
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch competitors"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	ctx.JSON(http.StatusOK, gin.H{
		"competitors": competitors,
		"pagination": gin.H{
			"current_page": pagination.Page,
			"page_size":    pagination.PageSize,
			"total_pages":  totalPages,
			"total_count":  total,
			"has_next":     pagination.Page < totalPages,
			"has_previous": pagination.Page > 1,
		},
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
