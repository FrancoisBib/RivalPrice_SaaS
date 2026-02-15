package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rivalprice/api-go/controllers"
	"github.com/rivalprice/api-go/services"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Initialize services
	userService := services.NewUserService(db)
	projectService := services.NewProjectService(db)
	competitorService := services.NewCompetitorService(db)
	monitoredPageService := services.NewMonitoredPageService(db)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	projectController := controllers.NewProjectController(projectService)
	competitorController := controllers.NewCompetitorController(competitorService)
	monitoredPageController := controllers.NewMonitoredPageController(monitoredPageService)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Users
		users := v1.Group("/users")
		{
			users.POST("", userController.CreateUser)
			users.GET("", userController.ListUsers)
			users.GET("/:id", userController.GetUser)
		}

		// Projects
		projects := v1.Group("/projects")
		{
			projects.POST("", projectController.CreateProject)
			projects.GET("", projectController.ListProjects)
			projects.GET("/:id", projectController.GetProject)
		}

		// Competitors
		competitors := v1.Group("/competitors")
		{
			competitors.POST("", competitorController.CreateCompetitor)
			competitors.GET("", competitorController.ListCompetitors)
			competitors.GET("/:id", competitorController.GetCompetitor)
		}

		// Monitored Pages
		monitoredPages := v1.Group("/monitored_pages")
		{
			monitoredPages.POST("", monitoredPageController.CreateMonitoredPage)
			monitoredPages.GET("", monitoredPageController.ListMonitoredPages)
			monitoredPages.GET("/:id", monitoredPageController.GetMonitoredPage)
		}
	}
}
