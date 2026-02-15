package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rivalprice/scraper-go/models"
)

var (
	db          *gorm.DB
	redisClient *redis.Client
	httpClient  *http.Client
)

func initDB() {
	dsn := "host=localhost user=rivalprice password=rivalprice_secret dbname=rivalprice port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
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
	// Common price patterns
	patterns := []string{
		`\$[\d,]+\.?\d*`,
		`USD\s*[\d,]+\.?\d*`,
		`€[\d,]+\.?\d*`,
		`EUR\s*[\d,]+\.?\d*`,
		`£[\d,]+\.?\d*`,
		`GBP\s*[\d,]+\.?\d*`,
		`data-price="([^"]+)"`,
		`class="[^"]*price[^"]*"[^>]*>[\s]*([^<]+)`,
		`"price"\s*:\s*"([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
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

	// Extract title from HTML
	titleRe := regexp.MustCompile(`<title>([^<]+)</title>`)
	titleMatch := titleRe.FindStringSubmatch(html)
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

func mustJson(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func main() {
	initDB()
	initRedis()
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
