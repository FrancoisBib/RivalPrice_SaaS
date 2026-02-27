package services

import (
	"errors"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

type CompetitorService struct {
	db *gorm.DB
}

func NewCompetitorService(db *gorm.DB) *CompetitorService {
	return &CompetitorService{db: db}
}

func (s *CompetitorService) CreateCompetitor(projectID uint, name, url string) (*models.Competitor, error) {
	// Verify project exists
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	competitor := models.Competitor{
		ProjectID: projectID,
		Name:      name,
		URL:       url,
	}

	if err := s.db.Create(&competitor).Error; err != nil {
		return nil, errors.New("failed to create competitor")
	}

	return &competitor, nil
}

func (s *CompetitorService) GetCompetitorByID(id uint) (*models.Competitor, error) {
	var competitor models.Competitor
	if err := s.db.Preload("Project").First(&competitor, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("competitor not found")
		}
		return nil, err
	}
	return &competitor, nil
}

func (s *CompetitorService) GetAllCompetitorsPaginated(offset, limit int) ([]models.Competitor, int64, error) {
	var competitors []models.Competitor
	var total int64
	
	if err := s.db.Model(&models.Competitor{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := s.db.Preload("Project").Offset(offset).Limit(limit).Find(&competitors).Error; err != nil {
		return nil, 0, err
	}
	return competitors, total, nil
}

func (s *CompetitorService) GetCompetitorsByProjectIDPaginated(projectID uint, offset, limit int) ([]models.Competitor, int64, error) {
	var competitors []models.Competitor
	var total int64
	
	if err := s.db.Model(&models.Competitor{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := s.db.Where("project_id = ?", projectID).Offset(offset).Limit(limit).Find(&competitors).Error; err != nil {
		return nil, 0, err
	}
	return competitors, total, nil
}

// Deprecated: Use GetAllCompetitorsPaginated instead
func (s *CompetitorService) GetAllCompetitors() ([]models.Competitor, error) {
	var competitors []models.Competitor
	if err := s.db.Preload("Project").Find(&competitors).Error; err != nil {
		return nil, err
	}
	return competitors, nil
}

// Deprecated: Use GetCompetitorsByProjectIDPaginated instead
func (s *CompetitorService) GetCompetitorsByProjectID(projectID uint) ([]models.Competitor, error) {
	var competitors []models.Competitor
	if err := s.db.Where("project_id = ?", projectID).Find(&competitors).Error; err != nil {
		return nil, err
	}
	return competitors, nil
}

func (s *CompetitorService) UpdateCompetitor(id uint, name, url string) (*models.Competitor, error) {
	var competitor models.Competitor
	if err := s.db.First(&competitor, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("competitor not found")
		}
		return nil, err
	}

	competitor.Name = name
	competitor.URL = url
	if err := s.db.Save(&competitor).Error; err != nil {
		return nil, errors.New("failed to update competitor")
	}

	return &competitor, nil
}

func (s *CompetitorService) DeleteCompetitor(id uint) error {
	var competitor models.Competitor
	if err := s.db.First(&competitor, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("competitor not found")
		}
		return err
	}

	// Delete related monitored pages first
	s.db.Where("competitor_id = ?", id).Delete(&models.MonitoredPage{})

	if err := s.db.Delete(&competitor).Error; err != nil {
		return errors.New("failed to delete competitor")
	}

	return nil
}
