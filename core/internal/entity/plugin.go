package entity

import (
	"time"

	"gorm.io/gorm"
)

// Plugin representa um plugin no sistema
type Plugin struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Version     string         `json:"version" gorm:"not null"`
	Description string         `json:"description"`
	Status      string         `json:"status" gorm:"default:'inactive'"`
	Config      string         `json:"config" gorm:"type:text"` // JSON config
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName define o nome da tabela no banco de dados
func (Plugin) TableName() string {
	return "plugins"
}
