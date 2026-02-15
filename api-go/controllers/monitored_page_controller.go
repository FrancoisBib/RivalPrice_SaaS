package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/services"
)

type MonitoredPageController struct {
	monitoredPageService *services.MonitoredPageService
}

func NewMonitoredPageController(monitoredPageService *services.MonitoredPageService) *MonitoredPageController {
	return &MonitoredPageController{monitoredPageService: monitoredPageService}
}

type CreateMonitoredPageRequest struct {
	CompetitorID uint   `json:"competitor_id" binding:"required"`
	PageType     string `json:"page_type" binding:"required"` // pricing or features
	URL          string `json:"url" binding:"required"`
	CSSSelector  string `json:"css_selector"`
}

// CreateMonitoredPage - POST /monitored_pages
func (c *MonitoredPageController) CreateMonitoredPage(ctx *gin.Context) {
	var req CreateMonitoredPageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monitoredPage, err := c.monitoredPageService.CreateMonitoredPage(req.CompetitorID, req.PageType, req.URL, req.CSSSelector)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":       "Monitored page created successfully",
		"monitored_page": monitoredPage,
	})
}

// ListMonitoredPages - GET /monitored_pages
func (c *MonitoredPageController) ListMonitoredPages(ctx *gin.Context) {
	// Optional: filter by competitor_id
	competitorID := ctx.Query("competitor_id")
	if competitorID != "" {
		cid, err := strconv.ParseUint(competitorID, 10, 32)
		if err == nil {
			monitoredPages, err := c.monitoredPageService.GetMonitoredPagesByCompetitorID(uint(cid))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitored pages"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"monitored_pages": monitoredPages})
			return
		}
	}

	monitoredPages, err := c.monitoredPageService.GetAllMonitoredPages()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitored pages"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"monitored_pages": monitoredPages,
	})
}

// GetMonitoredPage - GET /monitored_pages/:id
func (c *MonitoredPageController) GetMonitoredPage(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitored page ID"})
		return
	}

	monitoredPage, err := c.monitoredPageService.GetMonitoredPageByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"monitored_page": monitoredPage})
}
