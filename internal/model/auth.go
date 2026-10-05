package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"size:128;uniqueIndex" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	FullName  string    `gorm:"size:128" json:"full_name"`
	RoleID    uint      `gorm:"index" json:"role_id"`
	Role      *Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Category    string    `gorm:"size:64" json:"category"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
