package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	TotalCount   int64 `json:"total_count"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 100

// GetPaginationParams extracts pagination parameters from the request
func GetPaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// Paginate performs pagination on a GORM query
func Paginate(db *gorm.DB, params PaginationParams, dest interface{}) (*PaginatedResponse, error) {
	var totalCount int64
	
	// Clone the query to count total
	if err := db.Model(dest).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Get paginated results
	if err := db.Offset(params.Offset).Limit(params.PageSize).Find(dest).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(params.PageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		Data: dest,
		Pagination: PaginationMeta{
			CurrentPage: params.Page,
			PageSize:    params.PageSize,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			HasNext:     params.Page < totalPages,
			HasPrevious: params.Page > 1,
		},
	}, nil
}

// PaginateRaw performs pagination with raw query
func PaginateRaw(db *gorm.DB, params PaginationParams) *gorm.DB {
	return db.Offset(params.Offset).Limit(params.PageSize)
}
