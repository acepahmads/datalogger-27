package repository

import (
	"datalogger/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List returns all parameter categories, optionally filtering by enabled status, ordered by display_order
func (r *CategoryRepository) List(enabledOnly bool) ([]model.ParameterCategory, error) {
	return r.ListWithDomain(enabledOnly, "")
}

// ListWithDomain returns parameter categories, optionally filtered by enabled status and domain
func (r *CategoryRepository) ListWithDomain(enabledOnly bool, domain string) ([]model.ParameterCategory, error) {
	var categories []model.ParameterCategory
	query := r.db.Model(&model.ParameterCategory{}).Where("deleted_at IS NULL")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	if domain != "" && domain != "ALL" && domain != "CUSTOM" {
		query = query.Where("domain = ? OR domain = 'ALL' OR domain = 'CUSTOM' OR domain = '' OR domain IS NULL", domain)
	}
	err := query.Order("display_order ASC, id ASC").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// Calculate parameter counts per category
	type CountResult struct {
		CategoryID uint
		Count      int
	}
	var counts []CountResult
	_ = r.db.Model(&model.Parameter{}).
		Select("category_id, count(*) as count").
		Where("deleted_at IS NULL AND category_id IS NOT NULL").
		Group("category_id").
		Scan(&counts).Error

	countMap := make(map[uint]int)
	for _, c := range counts {
		countMap[c.CategoryID] = c.Count
	}

	for i := range categories {
		categories[i].ParameterCount = countMap[categories[i].ID]
	}

	return categories, nil
}

// GetByID finds a category by its ID
func (r *CategoryRepository) GetByID(id uint) (*model.ParameterCategory, error) {
	var cat model.ParameterCategory
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&cat).Error
	if err != nil {
		return nil, err
	}
	var count int64
	_ = r.db.Model(&model.Parameter{}).Where("category_id = ? AND deleted_at IS NULL", cat.ID).Count(&count).Error
	cat.ParameterCount = int(count)
	return &cat, nil
}

// GetByCode finds a category by its unique code
func (r *CategoryRepository) GetByCode(code string) (*model.ParameterCategory, error) {
	var cat model.ParameterCategory
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&cat).Error
	if err != nil {
		return nil, err
	}
	var count int64
	_ = r.db.Model(&model.Parameter{}).Where("category_id = ? AND deleted_at IS NULL", cat.ID).Count(&count).Error
	cat.ParameterCount = int(count)
	return &cat, nil
}

// Create inserts a new parameter category
func (r *CategoryRepository) Create(cat *model.ParameterCategory) error {
	return r.db.Create(cat).Error
}

// Update saves modifications to a category
func (r *CategoryRepository) Update(cat *model.ParameterCategory) error {
	return r.db.Save(cat).Error
}

// Delete safely deletes or archives a category, reassigning or nullifying referencing parameters
// Ensures parameter records, telemetry, alarms, and aggregations are NEVER deleted!
func (r *CategoryRepository) Delete(id uint, reassignCategoryID *uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if reassignCategoryID != nil && *reassignCategoryID > 0 {
			if err := tx.Model(&model.Parameter{}).
				Where("category_id = ?", id).
				Update("category_id", *reassignCategoryID).Error; err != nil {
				return err
			}
		} else {
			// Nullify foreign key so parameters remain accessible as Uncategorized
			if err := tx.Model(&model.Parameter{}).
				Where("category_id = ?", id).
				Update("category_id", nil).Error; err != nil {
				return err
			}
		}
		// Soft delete the category
		return tx.Delete(&model.ParameterCategory{}, id).Error
	})
}

// Reorder updates display_order for a list of category IDs
func (r *CategoryRepository) Reorder(categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for order, id := range categoryIDs {
			if err := tx.Model(&model.ParameterCategory{}).
				Where("id = ?", id).
				Update("display_order", order+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AssignParameters sets category_id for a list of parameters
func (r *CategoryRepository) AssignParameters(categoryID *uint, paramIDs []uint) error {
	if len(paramIDs) == 0 {
		return nil
	}
	return r.db.Model(&model.Parameter{}).
		Where("id IN ?", paramIDs).
		Update("category_id", categoryID).Error
}

// CountParameters counts active parameters in a category
func (r *CategoryRepository) CountParameters(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Parameter{}).Where("category_id = ? AND deleted_at IS NULL", categoryID).Count(&count).Error
	return count, err
}
