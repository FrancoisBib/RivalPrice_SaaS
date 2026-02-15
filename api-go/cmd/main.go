package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rivalprice/api-go/models"
	"github.com/rivalprice/api-go/routes"
)

var (
	db    *gorm.DB
	redisClient *redis.Client
)

func initDB() {
	dsn := "host=localhost user=rivalprice password=rivalprice_secret dbname=rivalprice port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated successfully")
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	log.Println("✅ Redis connected")
}

func main() {
	// Initialize connections
	initDB()
	initRedis()

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Database status endpoint
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

	// Redis status endpoint
	r.GET("/redis/status", func(c *gin.Context) {
		_, err := redisClient.Ping(context.Background()).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "disconnected", "redis": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "connected", "redis": "ok"})
	})

	// Migration status endpoint
	r.GET("/migrate", func(c *gin.Context) {
		tables := []string{"users", "projects", "competitors", "monitored_pages", "snapshots"}
		var existingTables []string
		
		for _, table := range tables {
			if db.Migrator().HasTable(table) {
				existingTables = append(existingTables, table)
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "migrated",
			"tables": existingTables,
		})
	})

	// Setup API routes
	routes.SetupRoutes(r, db)

	// Get port from env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server starting on port %s\n", port)
	r.Run(":" + port)
}
