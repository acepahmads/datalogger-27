package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"datalogger/internal/model"
	"datalogger/internal/repository"
)

var codeRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
	systemRepo   *repository.SystemRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository, systemRepo *repository.SystemRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		systemRepo:   systemRepo,
	}
}

type CreateCategoryRequest struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	IconKey           string `json:"icon_key"`
	ColorKey          string `json:"color_key"`
	DisplayOrder      int    `json:"display_order"`
	DefaultColumnSpan int    `json:"default_column_span"`
	Enabled           *bool  `json:"enabled"`
}

type UpdateCategoryRequest struct {
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	IconKey           *string `json:"icon_key"`
	ColorKey          *string `json:"color_key"`
	DisplayOrder      *int    `json:"display_order"`
	DefaultColumnSpan *int    `json:"default_column_span"`
	Enabled           *bool   `json:"enabled"`
}

type ReorderCategoriesRequest struct {
	CategoryIDs []uint `json:"category_ids"`
}

type AssignParametersRequest struct {
	CategoryID   *uint  `json:"category_id"`
	ParameterIDs []uint `json:"parameter_ids"`
}

func (s *CategoryService) ListCategories(enabledOnly bool) ([]model.ParameterCategory, error) {
	return s.categoryRepo.List(enabledOnly)
}

func (s *CategoryService) GetCategory(id uint) (*model.ParameterCategory, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *CategoryService) CreateCategory(req *CreateCategoryRequest, actor, ip, ua string) (*model.ParameterCategory, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, errors.New("category code is required")
	}
	if !codeRegex.MatchString(code) {
		return nil, errors.New("category code must contain only letters, numbers, underscores, and hyphens")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("category name is required")
	}

	// Check unique code
	if existing, _ := s.categoryRepo.GetByCode(code); existing != nil {
		return nil, fmt.Errorf("category with code '%s' already exists", code)
	}

	icon := strings.TrimSpace(req.IconKey)
	if icon == "" || !model.ValidCategoryIcons[icon] {
		icon = "activity"
	}

	color := strings.TrimSpace(req.ColorKey)
	if color == "" || !model.ValidCategoryColors[color] {
		color = "cyan"
	}

	span := req.DefaultColumnSpan
	if span < 1 || span > 4 {
		span = 1
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	cat := &model.ParameterCategory{
		Code:              code,
		Name:              name,
		Description:       strings.TrimSpace(req.Description),
		IconKey:           icon,
		ColorKey:          color,
		DisplayOrder:      req.DisplayOrder,
		DefaultColumnSpan: span,
		Enabled:           enabled,
	}

	if err := s.categoryRepo.Create(cat); err != nil {
		return nil, err
	}

	s.logAudit("CATEGORY_CREATED", fmt.Sprintf("ID:%d", cat.ID), actor, ip, ua,
		fmt.Sprintf("Created parameter category '%s' (Code: %s, Icon: %s, Color: %s)", cat.Name, cat.Code, cat.IconKey, cat.ColorKey))

	return cat, nil
}

func (s *CategoryService) UpdateCategory(id uint, req *UpdateCategoryRequest, actor, ip, ua string) (*model.ParameterCategory, error) {
	cat, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.New("category name cannot be empty")
		}
		cat.Name = name
	}

	if req.Description != nil {
		cat.Description = strings.TrimSpace(*req.Description)
	}

	if req.IconKey != nil {
		icon := strings.TrimSpace(*req.IconKey)
		if icon != "" && model.ValidCategoryIcons[icon] {
			cat.IconKey = icon
		}
	}

	if req.ColorKey != nil {
		color := strings.TrimSpace(*req.ColorKey)
		if color != "" && model.ValidCategoryColors[color] {
			cat.ColorKey = color
		}
	}

	if req.DisplayOrder != nil {
		cat.DisplayOrder = *req.DisplayOrder
	}

	if req.DefaultColumnSpan != nil {
		span := *req.DefaultColumnSpan
		if span >= 1 && span <= 4 {
			cat.DefaultColumnSpan = span
		}
	}

	if req.Enabled != nil {
		cat.Enabled = *req.Enabled
	}

	if err := s.categoryRepo.Update(cat); err != nil {
		return nil, err
	}

	s.logAudit("CATEGORY_UPDATED", fmt.Sprintf("ID:%d", cat.ID), actor, ip, ua,
		fmt.Sprintf("Updated parameter category '%s' (Enabled: %t, Span: %d)", cat.Name, cat.Enabled, cat.DefaultColumnSpan))

	return cat, nil
}

func (s *CategoryService) ToggleCategory(id uint, enabled bool, actor, ip, ua string) (*model.ParameterCategory, error) {
	cat, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("category not found")
	}
	cat.Enabled = enabled
	if err := s.categoryRepo.Update(cat); err != nil {
		return nil, err
	}

	s.logAudit("CATEGORY_TOGGLED", fmt.Sprintf("ID:%d", cat.ID), actor, ip, ua,
		fmt.Sprintf("Toggled parameter category '%s' enabled=%t", cat.Name, enabled))

	return cat, nil
}

func (s *CategoryService) DeleteCategory(id uint, reassignCategoryID *uint, actor, ip, ua string) error {
	cat, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return errors.New("category not found")
	}

	count, _ := s.categoryRepo.CountParameters(id)

	if err := s.categoryRepo.Delete(id, reassignCategoryID); err != nil {
		return err
	}

	reassignMsg := "unassigned (set to Uncategorized)"
	if reassignCategoryID != nil && *reassignCategoryID > 0 {
		reassignMsg = fmt.Sprintf("reassigned to category ID %d", *reassignCategoryID)
	}

	s.logAudit("CATEGORY_DELETED", fmt.Sprintf("ID:%d", id), actor, ip, ua,
		fmt.Sprintf("Deleted parameter category '%s' (Code: %s). %d referencing parameters %s. All parameters and telemetry retained.",
			cat.Name, cat.Code, count, reassignMsg))

	return nil
}

func (s *CategoryService) ReorderCategories(ids []uint, actor, ip, ua string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.categoryRepo.Reorder(ids); err != nil {
		return err
	}

	s.logAudit("CATEGORIES_REORDERED", "BULK", actor, ip, ua,
		fmt.Sprintf("Reordered %d parameter categories", len(ids)))

	return nil
}

func (s *CategoryService) AssignParameters(categoryID *uint, paramIDs []uint, actor, ip, ua string) error {
	if len(paramIDs) == 0 {
		return nil
	}

	var targetName = "Uncategorized"
	if categoryID != nil && *categoryID > 0 {
		cat, err := s.categoryRepo.GetByID(*categoryID)
		if err != nil {
			return errors.New("target category not found")
		}
		targetName = cat.Name
	}

	if err := s.categoryRepo.AssignParameters(categoryID, paramIDs); err != nil {
		return err
	}

	s.logAudit("PARAMETERS_ASSIGNED_CATEGORY", targetName, actor, ip, ua,
		fmt.Sprintf("Assigned %d parameters to category '%s'", len(paramIDs), targetName))

	return nil
}

func (s *CategoryService) logAudit(action, resource, username, ip, ua, details string) {
	if s.systemRepo != nil {
		_ = s.systemRepo.AddAuditTrail(&model.AuditTrail{
			Username:  username,
			Action:    action,
			Resource:  resource,
			IPAddress: ip,
			UserAgent: ua,
			Details:   details,
		})
	}
}
