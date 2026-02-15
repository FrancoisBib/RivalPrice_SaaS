package services

import (
	"errors"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

type MonitoredPageService struct {
	db *gorm.DB
}

func NewMonitoredPageService(db *gorm.DB) *MonitoredPageService {
	return &MonitoredPageService{db: db}
}

func (s *MonitoredPageService) CreateMonitoredPage(competitorID uint, pageType, url, cssSelector string) (*models.MonitoredPage, error) {
	// Verify competitor exists
	var competitor models.Competitor
	if err := s.db.First(&competitor, competitorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("competitor not found")
		}
		return nil, err
	}

	monitoredPage := models.MonitoredPage{
		CompetitorID: competitorID,
		PageType:     pageType,
		URL:          url,
		CSSSelector:  cssSelector,
	}

	if err := s.db.Create(&monitoredPage).Error; err != nil {
		return nil, errors.New("failed to create monitored page")
	}

	return &monitoredPage, nil
}

func (s *MonitoredPageService) GetMonitoredPageByID(id uint) (*models.MonitoredPage, error) {
	var monitoredPage models.MonitoredPage
	if err := s.db.Preload("Competitor").First(&monitoredPage, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("monitored page not found")
		}
		return nil, err
	}
	return &monitoredPage, nil
}

func (s *MonitoredPageService) GetAllMonitoredPages() ([]models.MonitoredPage, error) {
	var monitoredPages []models.MonitoredPage
	if err := s.db.Preload("Competitor").Find(&monitoredPages).Error; err != nil {
		return nil, err
	}
	return monitoredPages, nil
}

func (s *MonitoredPageService) GetMonitoredPagesByCompetitorID(competitorID uint) ([]models.MonitoredPage, error) {
	var monitoredPages []models.MonitoredPage
	if err := s.db.Where("competitor_id = ?", competitorID).Find(&monitoredPages).Error; err != nil {
		return nil, err
	}
	return monitoredPages, nil
}

func (s *MonitoredPageService) UpdateMonitoredPage(id uint, pageType, url, cssSelector string) (*models.MonitoredPage, error) {
	var monitoredPage models.MonitoredPage
	if err := s.db.First(&monitoredPage, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("monitored page not found")
		}
		return nil, err
	}

	monitoredPage.PageType = pageType
	monitoredPage.URL = url
	monitoredPage.CSSSelector = cssSelector

	if err := s.db.Save(&monitoredPage).Error; err != nil {
		return nil, errors.New("failed to update monitored page")
	}

	return &monitoredPage, nil
}

func (s *MonitoredPageService) DeleteMonitoredPage(id uint) error {
	var monitoredPage models.MonitoredPage
	if err := s.db.First(&monitoredPage, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("monitored page not found")
		}
		return err
	}

	// Delete related snapshots first
	s.db.Where("monitored_page_id = ?", id).Delete(&models.Snapshot{})

	if err := s.db.Delete(&monitoredPage).Error; err != nil {
		return errors.New("failed to delete monitored page")
	}

	return nil
}
