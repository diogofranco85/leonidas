package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DatabaseService gerencia a conexão com o banco de dados
type DatabaseService struct {
	db *gorm.DB
}

// CacheService gerencia a conexão com o cache Redis
type CacheService struct {
	// Será implementado quando integrarmos com o Redis do core
}

// NewDatabaseService cria uma nova instância do serviço de banco
func NewDatabaseService(db *gorm.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

// NewCacheService cria uma nova instância do serviço de cache
func NewCacheService() *CacheService {
	return &CacheService{}
}

// UserRepository gerencia operações de usuários
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository cria uma nova instância do repositório de usuários
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create cria um novo usuário
func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// GetByID obtém um usuário por ID
func (r *UserRepository) GetByID(id string) (*User, error) {
	var user User
	err := r.db.Preload("Roles.Permissions").Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetByUsername obtém um usuário por username
func (r *UserRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByEmail obtém um usuário por email
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Preload("Roles.Permissions").Where("email = ?", email).First(&user).Error
	return &user, err
}

// Update atualiza um usuário
func (r *UserRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete realiza soft delete de um usuário
func (r *UserRepository) Delete(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}

// List lista usuários com paginação
func (r *UserRepository) List(offset, limit int) ([]User, int64, error) {
	var users []User
	var total int64

	// Contar total
	if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar usuários
	err := r.db.Preload("Roles").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// Enable habilita um usuário
func (r *UserRepository) Enable(id string) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("is_enabled", true).Error
}

// Disable desabilita um usuário
func (r *UserRepository) Disable(id string) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("is_enabled", false).Error
}

// UpdateLastLogin atualiza o último login do usuário
func (r *UserRepository) UpdateLastLogin(id string) error {
	now := time.Now()
	return r.db.Model(&User{}).Where("id = ?", id).Update("last_login_at", now).Error
}

// RoleRepository gerencia operações de roles
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository cria uma nova instância do repositório de roles
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create cria uma nova role
func (r *RoleRepository) Create(role *Role) error {
	return r.db.Create(role).Error
}

// GetByID obtém uma role por ID
func (r *RoleRepository) GetByID(id string) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	return &role, err
}

// GetByName obtém uma role por nome
func (r *RoleRepository) GetByName(name string) (*Role, error) {
	var role Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

// List lista todas as roles
func (r *RoleRepository) List() ([]Role, error) {
	var roles []Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// Update atualiza uma role
func (r *RoleRepository) Update(role *Role) error {
	return r.db.Save(role).Error
}

// Delete realiza soft delete de uma role
func (r *RoleRepository) Delete(id string) error {
	return r.db.Delete(&Role{}, "id = ?", id).Error
}

// AddPermission adiciona uma permissão a uma role
func (r *RoleRepository) AddPermission(roleID, permissionID string) error {
	roleUUID, err := uuid.Parse(roleID)
	if err != nil {
		return fmt.Errorf("roleID inválido: %w", err)
	}

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return fmt.Errorf("permissionID inválido: %w", err)
	}

	rolePermission := RolePermission{
		RoleID:       roleUUID,
		PermissionID: permissionUUID,
		CreatedAt:    time.Now(),
	}
	return r.db.Create(&rolePermission).Error
}

// RemovePermission remove uma permissão de uma role
func (r *RoleRepository) RemovePermission(roleID, permissionID string) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&RolePermission{}).Error
}

// PermissionRepository gerencia operações de permissões
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository cria uma nova instância do repositório de permissões
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create cria uma nova permissão
func (r *PermissionRepository) Create(permission *Permission) error {
	return r.db.Create(permission).Error
}

// GetByID obtém uma permissão por ID
func (r *PermissionRepository) GetByID(id string) (*Permission, error) {
	var permission Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	return &permission, err
}

// GetByName obtém uma permissão por nome
func (r *PermissionRepository) GetByName(name string) (*Permission, error) {
	var permission Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	return &permission, err
}

// List lista todas as permissões
func (r *PermissionRepository) List() ([]Permission, error) {
	var permissions []Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

// Update atualiza uma permissão
func (r *PermissionRepository) Update(permission *Permission) error {
	return r.db.Save(permission).Error
}

// Delete realiza soft delete de uma permissão
func (r *PermissionRepository) Delete(id string) error {
	return r.db.Delete(&Permission{}, "id = ?", id).Error
}

// UserHistoryRepository gerencia operações de histórico de usuários
type UserHistoryRepository struct {
	db *gorm.DB
}

// NewUserHistoryRepository cria uma nova instância do repositório de histórico
func NewUserHistoryRepository(db *gorm.DB) *UserHistoryRepository {
	return &UserHistoryRepository{db: db}
}

// Create cria um novo registro de histórico
func (r *UserHistoryRepository) Create(history *UserHistory) error {
	return r.db.Create(history).Error
}

// GetByUserID obtém o histórico de um usuário
func (r *UserHistoryRepository) GetByUserID(userID string, offset, limit int) ([]UserHistory, int64, error) {
	var histories []UserHistory
	var total int64

	// Contar total
	if err := r.db.Model(&UserHistory{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar histórico
	err := r.db.Preload("CreatedByUser").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&histories).Error
	return histories, total, err
}

// GetRecent obtém o histórico recente do sistema
func (r *UserHistoryRepository) GetRecent(offset, limit int) ([]UserHistory, int64, error) {
	var histories []UserHistory
	var total int64

	// Contar total
	if err := r.db.Model(&UserHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar histórico
	err := r.db.Preload("User").Preload("CreatedByUser").Order("created_at DESC").Offset(offset).Limit(limit).Find(&histories).Error
	return histories, total, err
}

// PasswordResetRepository gerencia operações de reset de senha
type PasswordResetRepository struct {
	db *gorm.DB
}

// NewPasswordResetRepository cria uma nova instância do repositório de reset de senha
func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create cria um novo token de reset de senha
func (r *PasswordResetRepository) Create(reset *PasswordReset) error {
	return r.db.Create(reset).Error
}

// GetByToken obtém um reset por token
func (r *PasswordResetRepository) GetByToken(token string) (*PasswordReset, error) {
	var reset PasswordReset
	err := r.db.Preload("User").Where("token = ? AND used = false", token).First(&reset).Error
	return &reset, err
}

// MarkAsUsed marca um token como usado
func (r *PasswordResetRepository) MarkAsUsed(token string) error {
	return r.db.Model(&PasswordReset{}).Where("token = ?", token).Update("used", true).Error
}

// CleanExpired remove tokens expirados
func (r *PasswordResetRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&PasswordReset{}).Error
}

// UserSessionRepository gerencia operações de sessões de usuários
type UserSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository cria uma nova instância do repositório de sessões
func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

// Create cria uma nova sessão
func (r *UserSessionRepository) Create(session *UserSession) error {
	return r.db.Create(session).Error
}

// GetByToken obtém uma sessão por token
func (r *UserSessionRepository) GetByToken(token string) (*UserSession, error) {
	var session UserSession
	err := r.db.Preload("User").Where("token = ? AND is_active = true", token).First(&session).Error
	return &session, err
}

// GetByRefreshToken obtém uma sessão por refresh token
func (r *UserSessionRepository) GetByRefreshToken(refreshToken string) (*UserSession, error) {
	var session UserSession
	err := r.db.Preload("User").Where("refresh_token = ? AND is_active = true", refreshToken).First(&session).Error
	return &session, err
}

// Deactivate desativa uma sessão
func (r *UserSessionRepository) Deactivate(token string) error {
	return r.db.Model(&UserSession{}).Where("token = ?", token).Update("is_active", false).Error
}

// DeactivateAllUserSessions desativa todas as sessões de um usuário
func (r *UserSessionRepository) DeactivateAllUserSessions(userID string) error {
	return r.db.Model(&UserSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// CleanExpired remove sessões expiradas
func (r *UserSessionRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&UserSession{}).Error
}

// GetUserSessions obtém todas as sessões ativas de um usuário
func (r *UserSessionRepository) GetUserSessions(userID string) ([]UserSession, error) {
	var sessions []UserSession
	err := r.db.Where("user_id = ? AND is_active = true", userID).Find(&sessions).Error
	return sessions, err
}

// Cache operations (serão implementadas quando integrarmos com Redis)

// SetCache define um valor no cache
func (c *CacheService) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Será implementado quando integrarmos com Redis do core
	log.Printf("Cache SET: %s = %v (expires in %v)", key, value, expiration)
	return nil
}

// GetCache obtém um valor do cache
func (c *CacheService) GetCache(ctx context.Context, key string) (string, error) {
	// Será implementado quando integrarmos com Redis do core
	log.Printf("Cache GET: %s", key)
	return "", fmt.Errorf("cache não implementado ainda")
}

// DeleteCache remove um valor do cache
func (c *CacheService) DeleteCache(ctx context.Context, key string) error {
	// Será implementado quando integrarmos com Redis do core
	log.Printf("Cache DELETE: %s", key)
	return nil
}

// InvalidateUserCache invalida o cache de um usuário
func (c *CacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	keys := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("user_sessions:%s", userID),
		fmt.Sprintf("user_permissions:%s", userID),
	}

	for _, key := range keys {
		if err := c.DeleteCache(ctx, key); err != nil {
			log.Printf("Aviso: Falha ao invalidar cache %s: %v", key, err)
		}
	}

	return nil
}
