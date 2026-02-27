package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rivalprice/api-go/config"
	"github.com/rivalprice/api-go/middleware"
	"github.com/rivalprice/api-go/models"
	"github.com/rivalprice/api-go/routes"
	"github.com/rivalprice/api-go/services"
	"github.com/rivalprice/api-go/workers"
)

var (
	db            *gorm.DB
	redisClient   *redis.Client
	scrapingSvc   *services.ScrapingService
	schedulerSvc  *services.SchedulerService
	alertWorker   *workers.AlertWorker
	appConfig     *config.Config
)

func initDB(cfg *config.Config) {
	var err error
	db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ PostgreSQL connected")

	// AutoMigrate - create tables automatically
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Competitor{},
		&models.MonitoredPage{},
		&models.Snapshot{},
		&models.AlertLog{},
		&models.UserNotificationSettings{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated successfully")
}

func initRedis(cfg *config.Config) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	
	// Test connection
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	
	log.Println("✅ Redis connected")
}

func initServices() {
	scrapingSvc = services.NewScrapingService(db, redisClient)
	schedulerSvc = services.NewSchedulerService(db, redisClient)
}

func main() {
	// Load configuration
	appConfig = config.Load()
	if err := appConfig.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	log.Printf("🚀 Starting RivalPrice API in %s mode", appConfig.Environment)

	// Initialize connections
	initDB(appConfig)
	initRedis(appConfig)
	initServices()

	// Initialize scheduler for existing pages
	if err := schedulerSvc.InitializeScheduledPages(); err != nil {
		log.Printf("⚠️  Scheduler: failed to initialize pages: %v", err)
	}

	// Start scheduler in background
	go schedulerSvc.Start()

	// Start alert worker in background (polls every 30s)
	alertWorker = workers.NewAlertWorker(db)
	go alertWorker.Start()

	// Setup Gin
	if appConfig.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.StandardRateLimit())       // Rate limiting: 100 req/min
	r.Use(middleware.JWTAuthMiddleware(appConfig.JWTSecret)) // JWT authentication

	// Health check endpoint (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    appConfig.Environment,
		})
	})

	// Database status endpoint (public)
	r.GET("/db/status", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "disconnected", "database": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "connected", "database": "postgres"})
	})

	// Redis status endpoint (public)
	r.GET("/redis/status", func(c *gin.Context) {
		_, err := redisClient.Ping(context.Background()).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "disconnected", "redis": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "connected", "redis": "ok"})
	})

	// Migration status endpoint (public)
	r.GET("/migrate", func(c *gin.Context) {
		tables := []string{"users", "projects", "competitors", "monitored_pages", "snapshots", "alert_logs", "user_notification_settings"}
		var existingTables []string
		
		for _, table := range tables {
			if db.Migrator().HasTable(table) {
				existingTables = append(existingTables, table)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "migrated",
			"tables":  existingTables,
			"count":   len(existingTables),
		})
	})

	// Scrape endpoints with stricter rate limiting
	scrapeGroup := r.Group("/scrape")
	scrapeGroup.Use(middleware.StrictRateLimit()) // 10 req/min for scraping
	{
		scrapeGroup.POST("/page/:id", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
				return
			}

			if err := scrapingSvc.QueueScrapeJob(uint(id)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Scrape job queued", "page_id": id})
		})

		scrapeGroup.POST("/project/:id", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}

			if err := scrapingSvc.QueueScrapeJobForProject(uint(id)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Scrape jobs queued for project", "project_id": id})
		})
	}

	// Setup API routes
	routes.SetupRoutes(r, db, appConfig.JWTSecret)

	fmt.Printf("🚀 Server starting on port %s\n", appConfig.Port)
	if err := r.Run(":" + appConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
