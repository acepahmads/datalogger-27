package model

import (
	"time"

	"gorm.io/gorm"
)

// ParameterCategory represents a configurable domain category for telemetry parameters (e.g. Gas & Pollutants, Meteorology, Water Quality)
type ParameterCategory struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Code              string         `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name              string         `gorm:"size:128;not null" json:"name"`
	Description       string         `gorm:"size:255" json:"description"`
	IconKey           string         `gorm:"size:64;default:'activity'" json:"icon_key"`
	ColorKey          string         `gorm:"size:64;default:'cyan'" json:"color_key"`
	DisplayOrder      int            `gorm:"default:0;index" json:"display_order"`
	DefaultColumnSpan int            `gorm:"default:1" json:"default_column_span"`
	Enabled           bool           `gorm:"default:true;index" json:"enabled"`
	Domain            string         `gorm:"size:32;default:'ALL';index" json:"domain"`
	ParameterCount    int            `gorm:"-" json:"parameter_count,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// Monitoring domain constants
const (
	DomainAQMS   = "AQMS"
	DomainWQMS   = "WQMS"
	DomainCustom = "CUSTOM"
	DomainAll    = "ALL"
)

// ValidCategoryDomains defines permitted domain tags
var ValidCategoryDomains = map[string]bool{
	DomainAQMS:   true,
	DomainWQMS:   true,
	DomainCustom: true,
	DomainAll:    true,
}

// IsValidCategoryDomain checks whether the given domain tag is valid
func IsValidCategoryDomain(domain string) bool {
	return ValidCategoryDomains[domain]
}

// ValidCategoryIcons defines the permitted safe vector icon catalog
var ValidCategoryIcons = map[string]bool{
	"particles":   true,
	"atom":        true,
	"thermometer": true,
	"droplet":     true,
	"gauge":       true,
	"wind":        true,
	"compass":     true,
	"cloud-rain":  true,
	"sun":         true,
	"waves":       true,
	"activity":    true,
}

// ValidCategoryColors defines the permitted SCADA theme color tokens
var ValidCategoryColors = map[string]bool{
	"cyan":    true,
	"blue":    true,
	"amber":   true,
	"emerald": true,
	"purple":  true,
	"rose":    true,
	"teal":    true,
	"sky":     true,
	"indigo":  true,
	"slate":   true,
}
