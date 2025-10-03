package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Migration representa uma migration executada
type Migration struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string         `gorm:"uniqueIndex;not null" json:"name"`
	ExecutedAt time.Time      `gorm:"default:now()" json:"executedAt"`
	CreatedAt  time.Time      `gorm:"default:now()" json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// User representa um usuário do sistema
type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Username    string         `gorm:"uniqueIndex;not null" json:"username"`
	Email       string         `gorm:"uniqueIndex;not null" json:"email"`
	Password    string         `gorm:"not null" json:"-"`
	FirstName   string         `json:"firstName"`
	LastName    string         `json:"lastName"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	IsEnabled   bool           `gorm:"default:true" json:"isEnabled"`
	LastLoginAt *time.Time     `json:"lastLoginAt"`
	CreatedAt   time.Time      `gorm:"default:now()" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"default:now()" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	Roles       []Role        `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	UserHistory []UserHistory `gorm:"foreignKey:UserID" json:"history,omitempty"`
}

// Role representa uma role/perfil do sistema
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"default:now()" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"default:now()" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission representa uma permissão do sistema
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	Resource    string         `gorm:"not null" json:"resource"` // Ex: "users", "roles", "permissions"
	Action      string         `gorm:"not null" json:"action"`   // Ex: "create", "read", "update", "delete"
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"default:now()" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"default:now()" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// UserHistory representa o histórico de alterações de um usuário
type UserHistory struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"not null;index" json:"userId"`
	Action      string         `gorm:"not null" json:"action"` // Ex: "created", "updated", "deleted", "enabled", "disabled"
	Field       string         `json:"field"`                  // Campo alterado (opcional)
	OldValue    string         `json:"oldValue"`               // Valor anterior (opcional)
	NewValue    string         `json:"newValue"`               // Novo valor (opcional)
	Description string         `json:"description"`            // Descrição da alteração
	IPAddress   string         `json:"ipAddress"`              // IP de onde veio a alteração
	UserAgent   string         `json:"userAgent"`              // User agent do cliente
	CreatedBy   *uuid.UUID     `json:"createdBy"`              // ID do usuário que fez a alteração
	CreatedAt   time.Time      `gorm:"default:now()" json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	User          User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedByUser *User `gorm:"foreignKey:CreatedBy" json:"createdByUser,omitempty"`
}

// PasswordReset representa um token de reset de senha
type PasswordReset struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"not null;index" json:"userId"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null" json:"expiresAt"`
	Used      bool           `gorm:"default:false" json:"used"`
	CreatedAt time.Time      `gorm:"default:now()" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserSession representa uma sessão ativa do usuário
type UserSession struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `gorm:"not null;index" json:"userId"`
	Token            string         `gorm:"uniqueIndex;not null" json:"token"`
	RefreshToken     string         `gorm:"uniqueIndex;not null" json:"refreshToken"`
	ExpiresAt        time.Time      `gorm:"not null" json:"expiresAt"`
	RefreshExpiresAt time.Time      `gorm:"not null" json:"refreshExpiresAt"`
	IPAddress        string         `json:"ipAddress"`
	UserAgent        string         `json:"userAgent"`
	IsActive         bool           `gorm:"default:true" json:"isActive"`
	CreatedAt        time.Time      `gorm:"default:now()" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"default:now()" json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamentos
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Tabelas de relacionamento many-to-many
type UserRole struct {
	UserID    uuid.UUID `gorm:"primaryKey" json:"userId"`
	RoleID    uuid.UUID `gorm:"primaryKey" json:"roleId"`
	CreatedAt time.Time `gorm:"default:now()" json:"createdAt"`
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"primaryKey" json:"roleId"`
	PermissionID uuid.UUID `gorm:"primaryKey" json:"permissionId"`
	CreatedAt    time.Time `gorm:"default:now()" json:"createdAt"`
}

// Métodos auxiliares para as entidades

// GetFullName retorna o nome completo do usuário
func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}

// IsAdmin verifica se o usuário tem a role de administrador
func (u *User) IsAdmin() bool {
	for _, role := range u.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}

// HasPermission verifica se o usuário tem uma permissão específica
func (u *User) HasPermission(resource, action string) bool {
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if permission.Resource == resource && permission.Action == action {
				return true
			}
		}
	}
	return false
}

// HasRole verifica se o usuário tem uma role específica
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// IsExpired verifica se o token de reset de senha expirou
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// IsExpired verifica se a sessão expirou
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// IsRefreshExpired verifica se o refresh token expirou
func (us *UserSession) IsRefreshExpired() bool {
	return time.Now().After(us.RefreshExpiresAt)
}
