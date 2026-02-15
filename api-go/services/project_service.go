package services

import (
	"errors"

	"github.com/rivalprice/api-go/models"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) CreateProject(userID uint, name string) (*models.Project, error) {
	// Verify user exists
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	project := models.Project{
		UserID: userID,
		Name:   name,
	}

	if err := s.db.Create(&project).Error; err != nil {
		return nil, errors.New("failed to create project")
	}

	return &project, nil
}

func (s *ProjectService) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	if err := s.db.Preload("User").First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return &project, nil
}

func (s *ProjectService) GetAllProjects() ([]models.Project, error) {
	var projects []models.Project
	if err := s.db.Preload("User").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectService) GetProjectsByUserID(userID uint) ([]models.Project, error) {
	var projects []models.Project
	if err := s.db.Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectService) UpdateProject(id uint, name string) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	project.Name = name
	if err := s.db.Save(&project).Error; err != nil {
		return nil, errors.New("failed to update project")
	}

	return &project, nil
}

func (s *ProjectService) DeleteProject(id uint) error {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		return err
	}

	// Delete related competitors and monitored pages first
	s.db.Where("project_id = ?", id).Delete(&models.Competitor{})
	
	if err := s.db.Delete(&project).Error; err != nil {
		return errors.New("failed to delete project")
	}

	return nil
}
