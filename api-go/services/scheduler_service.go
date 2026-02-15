package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/rivalprice/api-go/models"
)

type SchedulerService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewSchedulerService(db *gorm.DB, redisClient *redis.Client) *SchedulerService {
	return &SchedulerService{
		db:    db,
		redis: redisClient,
	}
}

func (s *SchedulerService) Start() {
	log.Println("⏰ Scheduler started, ticking every 60 seconds")
	
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *SchedulerService) tick() {
	now := time.Now()

	var pages []models.MonitoredPage
	if err := s.db.Where("next_run_at <= ?", now).Find(&pages).Error; err != nil {
		log.Printf("❌ Scheduler: failed to query pages: %v", err)
		return
	}

	if len(pages) == 0 {
		return
	}

	log.Printf("📅 Scheduler: found %d pages to scrape", len(pages))

	for _, page := range pages {
		if err := s.queuePage(page); err != nil {
			log.Printf("❌ Scheduler: failed to queue page %d: %v", page.ID, err)
			continue
		}
		log.Printf("✅ Scheduler: queued page %d (%s)", page.ID, page.URL)
	}
}

func (s *SchedulerService) queuePage(page models.MonitoredPage) error {
	job := ScrapeJob{
		PageID: page.ID,
		URL:    page.URL,
		Type:   page.PageType,
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Push to Redis queue with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.redis.RPush(ctx, "scrape_job", jobJSON).Err(); err != nil {
		return err
	}

	// Update last_checked_at
	now := time.Now()
	if err := s.db.Model(&page).Updates(map[string]interface{}{
		"last_checked_at": now,
		"next_run_at":     s.calculateNextRun(page.Frequency),
	}).Error; err != nil {
		return err
	}

	return nil
}

func (s *SchedulerService) calculateNextRun(frequency models.Frequency) time.Time {
	now := time.Now()

	switch frequency {
	case models.FrequencyDaily:
		return now.Add(24 * time.Hour)
	case models.FrequencyWeekly:
		return now.Add(7 * 24 * time.Hour)
	case models.FrequencyMonthly:
		return now.Add(30 * 24 * time.Hour)
	default:
		return now.Add(24 * time.Hour)
	}
}

// InitializeScheduledPages sets next_run_at for pages that don't have it
func (s *SchedulerService) InitializeScheduledPages() error {
	now := time.Now()
	
	return s.db.Model(&models.MonitoredPage{}).
		Where("next_run_at IS NULL").
		Updates(map[string]interface{}{
			"next_run_at": now,
		}).Error
}
