package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rivalprice/scraper-go/models"
)

var (
	db           *gorm.DB
	redisClient  *redis.Client
	httpClient   *http.Client
	
	// Pre-compiled regex patterns for price extraction
	pricePatterns = []*regexp.Regexp{
		regexp.MustCompile(`\$[\d,]+\.?\d*`),
		regexp.MustCompile(`USD\s*[\d,]+\.?\d*`),
		regexp.MustCompile(`€[\d,]+\.?\d*`),
		regexp.MustCompile(`EUR\s*[\d,]+\.?\d*`),
		regexp.MustCompile(`£[\d,]+\.?\d*`),
		regexp.MustCompile(`GBP\s*[\d,]+\.?\d*`),
		regexp.MustCompile(`data-price="([^"]+)"`),
		regexp.MustCompile(`class="[^"]*price[^"]*"[^>]*>[\s]*([^<]+)`),
		regexp.MustCompile(`"price"\s*:\s*"([^"]+)"`),
	}
	
	// Pre-compiled regex for title extraction
	titleRegex = regexp.MustCompile(`<title>([^<]+)</title>`)
)

// Config holds scraper configuration
type Config struct {
	DatabaseURL string
	RedisAddr   string
	RedisPass   string
	RedisDB     int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=rivalprice password=rivalprice_secret dbname=rivalprice port=5432 sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:     getEnvAsInt("REDIS_DB", 0),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func initDB(cfg *Config) {
	var err error
	db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ PostgreSQL connected")

	err = db.AutoMigrate(&models.MonitoredPage{}, &models.Snapshot{})
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
	log.Println("✅ Database migrated")
}

func initRedis(cfg *Config) {
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

func initHTTP() {
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	log.Println("✅ HTTP client ready")
}

type ScrapeJob struct {
	PageID uint   `json:"page_id"`
	URL    string `json:"url"`
	Type   string `json:"type"`
}

func extractPrice(html string) string {
	for _, re := range pricePatterns {
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
		if matches != nil {
			return strings.TrimSpace(matches[0])
		}
	}
	return ""
}

func extractAvailability(html string) string {
	htmlLower := strings.ToLower(html)
	
	if strings.Contains(htmlLower, "out of stock") || strings.Contains(htmlLower, "outofstock") {
		return "out_of_stock"
	}
	if strings.Contains(htmlLower, "in stock") || strings.Contains(htmlLower, "instock") || strings.Contains(htmlLower, "available") {
		return "in_stock"
	}
	if strings.Contains(htmlLower, "pre-order") || strings.Contains(htmlLower, "preorder") {
		return "pre_order"
	}
	
	return "available"
}

// mustJson marshals v to JSON, returns empty bytes on error (with logging)
func mustJson(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("❌ Failed to marshal JSON: %v", err)
		return []byte("{}")
	}
	return data
}

func processScrapeJob(ctx context.Context, job ScrapeJob) error {
	log.Printf("🔄 Processing scrape job for page %d: %s", job.PageID, job.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	html := string(htmlBytes)

	price := extractPrice(html)
	availability := extractAvailability(html)

	// Extract title from HTML using pre-compiled regex
	titleMatch := titleRegex.FindStringSubmatch(html)
	title := ""
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	}

	rawData := map[string]interface{}{
		"title":        title,
		"url":          job.URL,
		"html":         html,
		"price_found":  price,
		"availability": availability,
		"status_code":  resp.StatusCode,
	}

	snapshot := models.Snapshot{
		MonitoredPageID: job.PageID,
		Price:          price,
		Availability:   availability,
		RawData:        mustJson(rawData),
		ScrapedAt:      time.Now(),
	}

	if err := db.Create(&snapshot).Error; err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}

	log.Printf("✅ Snapshot stored: Page %d, Price: %s, Availability: %s", job.PageID, price, availability)
	return nil
}

func main() {
	cfg := LoadConfig()
	
	initDB(cfg)
	initRedis(cfg)
	initHTTP()

	log.Println("🚀 Worker started, waiting for jobs...")

	for {
		result, err := redisClient.BRPop(context.Background(), 5*time.Second, "scrape_job").Result()
		if err != nil {
			continue
		}

		if len(result) < 2 {
			continue
		}

		var job ScrapeJob
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			log.Printf("❌ Failed to parse job: %v", err)
			continue
		}

		if err := processScrapeJob(context.Background(), job); err != nil {
			log.Printf("❌ Job failed: %v", err)
		}
	}
}
