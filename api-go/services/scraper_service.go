package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

type ScrapingService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewScrapingService(db *gorm.DB, redisClient *redis.Client) *ScrapingService {
	return &ScrapingService{
		db:    db,
		redis: redisClient,
	}
}

type ScrapeJob struct {
	PageID uint   `json:"page_id"`
	URL    string `json:"url"`
	Type   string `json:"type"`
}

// QueueScrapeJob adds a scraping job to the Redis queue
func (s *ScrapingService) QueueScrapeJob(pageID uint) error {
	// Get the monitored page
	var page models.MonitoredPage
	if err := s.db.First(&page, pageID).Error; err != nil {
		return err
	}

	job := ScrapeJob{
		PageID: page.ID,
		URL:    page.URL,
		Type:   page.PageType,
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Push to Redis queue
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.redis.RPush(ctx, "scrape_job", jobData).Err()
}

// QueueScrapeJobForProject queues scraping jobs for all pages in a project
func (s *ScrapingService) QueueScrapeJobForProject(projectID uint) error {
	var pages []models.MonitoredPage
	
	// Get all competitors for the project
	var competitors []models.Competitor
	if err := s.db.Where("project_id = ?", projectID).Find(&competitors).Error; err != nil {
		return err
	}

	// Get all monitored pages for these competitors
	competitorIDs := make([]uint, len(competitors))
	for i, c := range competitors {
		competitorIDs[i] = c.ID
	}

	if err := s.db.Where("competitor_id IN ?", competitorIDs).Find(&pages).Error; err != nil {
		return err
	}

	// Queue each page
	for _, page := range pages {
		if err := s.QueueScrapeJob(page.ID); err != nil {
			return err
		}
	}

	return nil
}
