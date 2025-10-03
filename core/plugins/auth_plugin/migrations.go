package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MigrationFunc representa uma função de migration
type MigrationFunc func(db *gorm.DB) error

// MigrationDefinition representa uma migration
type MigrationDefinition struct {
	Name string
	Func MigrationFunc
}

// getMigrations retorna todas as migrations do plugin
func (p *AuthPlugin) getMigrations() []MigrationDefinition {
	return []MigrationDefinition{
		{
			Name: "001_create_migrations_table",
			Func: p.migration001CreateMigrationsTable,
		},
		{
			Name: "002_create_users_table",
			Func: p.migration002CreateUsersTable,
		},
		{
			Name: "003_create_roles_table",
			Func: p.migration003CreateRolesTable,
		},
		{
			Name: "004_create_permissions_table",
			Func: p.migration004CreatePermissionsTable,
		},
		{
			Name: "005_create_user_roles_table",
			Func: p.migration005CreateUserRolesTable,
		},
		{
			Name: "006_create_role_permissions_table",
			Func: p.migration006CreateRolePermissionsTable,
		},
		{
			Name: "007_create_user_history_table",
			Func: p.migration007CreateUserHistoryTable,
		},
		{
			Name: "008_create_password_resets_table",
			Func: p.migration008CreatePasswordResetsTable,
		},
		{
			Name: "009_create_user_sessions_table",
			Func: p.migration009CreateUserSessionsTable,
		},
		{
			Name: "010_create_default_roles_and_permissions",
			Func: p.migration010CreateDefaultRolesAndPermissions,
		},
		{
			Name: "011_create_default_admin_user",
			Func: p.migration011CreateDefaultAdminUser,
		},
	}
}

// runMigrationsFromFile executa todas as migrations pendentes
func (p *AuthPlugin) runMigrationsFromFile(ctx context.Context) error {
	log.Printf("[%s] Iniciando execução de migrations...", p.info.Name)

	// Obter conexão com o banco (será implementado quando integrarmos com o core)
	// Por enquanto, vamos simular
	db := p.getDatabaseConnection()
	if db == nil {
		return fmt.Errorf("falha ao obter conexão com o banco de dados")
	}

	migrations := p.getMigrations()
	executedCount := 0

	for _, migration := range migrations {
		// Verificar se a migration já foi executada
		executed, err := p.isMigrationExecuted(db, migration.Name)
		if err != nil {
			return fmt.Errorf("falha ao verificar migration %s: %w", migration.Name, err)
		}

		if executed {
			log.Printf("[%s] Migration %s já foi executada, pulando...", p.info.Name, migration.Name)
			continue
		}

		// Executar a migration
		log.Printf("[%s] Executando migration: %s", p.info.Name, migration.Name)
		if err := migration.Func(db); err != nil {
			return fmt.Errorf("falha ao executar migration %s: %w", migration.Name, err)
		}

		// Registrar a migration como executada
		if err := p.recordMigration(db, migration.Name); err != nil {
			return fmt.Errorf("falha ao registrar migration %s: %w", migration.Name, err)
		}

		executedCount++
		log.Printf("[%s] Migration %s executada com sucesso", p.info.Name, migration.Name)
	}

	log.Printf("[%s] Migrations concluídas. %d migrations executadas", p.info.Name, executedCount)
	return nil
}

// isMigrationExecuted verifica se uma migration já foi executada
func (p *AuthPlugin) isMigrationExecuted(db *gorm.DB, migrationName string) (bool, error) {
	// Verificar se a tabela migrations existe
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'migrations'").Scan(&count).Error
	if err != nil {
		return false, err
	}

	// Se a tabela não existe, a migration não foi executada
	if count == 0 {
		return false, nil
	}

	// Verificar se a migration específica foi executada
	err = db.Model(&Migration{}).Where("name = ?", migrationName).Count(&count).Error
	return count > 0, err
}

// recordMigration registra uma migration como executada
func (p *AuthPlugin) recordMigration(db *gorm.DB, migrationName string) error {
	migration := Migration{
		ID:         uuid.New(),
		Name:       migrationName,
		ExecutedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
	return db.Create(&migration).Error
}

// Implementações das migrations

// migration001CreateMigrationsTable cria a tabela de migrations
func (p *AuthPlugin) migration001CreateMigrationsTable(db *gorm.DB) error {
	return db.AutoMigrate(&Migration{})
}

// migration002CreateUsersTable cria a tabela de usuários
func (p *AuthPlugin) migration002CreateUsersTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// migration003CreateRolesTable cria a tabela de roles
func (p *AuthPlugin) migration003CreateRolesTable(db *gorm.DB) error {
	return db.AutoMigrate(&Role{})
}

// migration004CreatePermissionsTable cria a tabela de permissões
func (p *AuthPlugin) migration004CreatePermissionsTable(db *gorm.DB) error {
	return db.AutoMigrate(&Permission{})
}

// migration005CreateUserRolesTable cria a tabela de relacionamento user-roles
func (p *AuthPlugin) migration005CreateUserRolesTable(db *gorm.DB) error {
	return db.AutoMigrate(&UserRole{})
}

// migration006CreateRolePermissionsTable cria a tabela de relacionamento role-permissions
func (p *AuthPlugin) migration006CreateRolePermissionsTable(db *gorm.DB) error {
	return db.AutoMigrate(&RolePermission{})
}

// migration007CreateUserHistoryTable cria a tabela de histórico de usuários
func (p *AuthPlugin) migration007CreateUserHistoryTable(db *gorm.DB) error {
	return db.AutoMigrate(&UserHistory{})
}

// migration008CreatePasswordResetsTable cria a tabela de reset de senhas
func (p *AuthPlugin) migration008CreatePasswordResetsTable(db *gorm.DB) error {
	return db.AutoMigrate(&PasswordReset{})
}

// migration009CreateUserSessionsTable cria a tabela de sessões de usuários
func (p *AuthPlugin) migration009CreateUserSessionsTable(db *gorm.DB) error {
	return db.AutoMigrate(&UserSession{})
}

// migration010CreateDefaultRolesAndPermissions cria roles e permissões padrão
func (p *AuthPlugin) migration010CreateDefaultRolesAndPermissions(db *gorm.DB) error {
	// Criar roles padrão
	roles := []Role{
		{
			ID:          uuid.New(),
			Name:        "admin",
			Description: "Administrador do sistema com acesso total",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "user",
			Description: "Usuário comum do sistema",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "moderator",
			Description: "Moderador com permissões limitadas",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Criar as roles
	for _, role := range roles {
		if err := db.FirstOrCreate(&role, Role{Name: role.Name}).Error; err != nil {
			return fmt.Errorf("falha ao criar role %s: %w", role.Name, err)
		}
	}

	// Criar permissões padrão
	permissions := []Permission{
		// Permissões de usuários
		{ID: uuid.New(), Name: "users:create", Description: "Criar usuários", Resource: "users", Action: "create", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "users:read", Description: "Visualizar usuários", Resource: "users", Action: "read", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "users:update", Description: "Atualizar usuários", Resource: "users", Action: "update", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "users:delete", Description: "Deletar usuários", Resource: "users", Action: "delete", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		// Permissões de roles
		{ID: uuid.New(), Name: "roles:create", Description: "Criar roles", Resource: "roles", Action: "create", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "roles:read", Description: "Visualizar roles", Resource: "roles", Action: "read", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "roles:update", Description: "Atualizar roles", Resource: "roles", Action: "update", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "roles:delete", Description: "Deletar roles", Resource: "roles", Action: "delete", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		// Permissões de permissões
		{ID: uuid.New(), Name: "permissions:create", Description: "Criar permissões", Resource: "permissions", Action: "create", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "permissions:read", Description: "Visualizar permissões", Resource: "permissions", Action: "read", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "permissions:update", Description: "Atualizar permissões", Resource: "permissions", Action: "update", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "permissions:delete", Description: "Deletar permissões", Resource: "permissions", Action: "delete", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		// Permissões de sistema
		{ID: uuid.New(), Name: "system:admin", Description: "Acesso administrativo total", Resource: "system", Action: "admin", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "system:logs", Description: "Visualizar logs do sistema", Resource: "system", Action: "logs", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// Criar as permissões
	for _, permission := range permissions {
		if err := db.FirstOrCreate(&permission, Permission{Name: permission.Name}).Error; err != nil {
			return fmt.Errorf("falha ao criar permissão %s: %w", permission.Name, err)
		}
	}

	// Associar permissões às roles
	// Admin tem todas as permissões
	var adminRole Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("falha ao encontrar role admin: %w", err)
	}

	for _, permission := range permissions {
		var perm Permission
		if err := db.Where("name = ?", permission.Name).First(&perm).Error; err != nil {
			continue // Pular se não encontrar a permissão
		}

		// Verificar se a associação já existe
		var count int64
		db.Model(&RolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).Count(&count)
		if count == 0 {
			rolePermission := RolePermission{
				RoleID:       adminRole.ID,
				PermissionID: perm.ID,
				CreatedAt:    time.Now(),
			}
			if err := db.Create(&rolePermission).Error; err != nil {
				log.Printf("Aviso: Falha ao associar permissão %s à role admin: %v", permission.Name, err)
			}
		}
	}

	// User tem permissões básicas
	var userRole Role
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		return fmt.Errorf("falha ao encontrar role user: %w", err)
	}

	userPermissions := []string{"users:read", "permissions:read"}
	for _, permName := range userPermissions {
		var perm Permission
		if err := db.Where("name = ?", permName).First(&perm).Error; err != nil {
			continue
		}

		var count int64
		db.Model(&RolePermission{}).Where("role_id = ? AND permission_id = ?", userRole.ID, perm.ID).Count(&count)
		if count == 0 {
			rolePermission := RolePermission{
				RoleID:       userRole.ID,
				PermissionID: perm.ID,
				CreatedAt:    time.Now(),
			}
			if err := db.Create(&rolePermission).Error; err != nil {
				log.Printf("Aviso: Falha ao associar permissão %s à role user: %v", permName, err)
			}
		}
	}

	log.Printf("[%s] Roles e permissões padrão criadas com sucesso", p.info.Name)
	return nil
}

// migration011CreateDefaultAdminUser cria o usuário administrador padrão
func (p *AuthPlugin) migration011CreateDefaultAdminUser(db *gorm.DB) error {
	// Verificar se já existe um usuário admin
	var count int64
	db.Model(&User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		log.Printf("[%s] Usuário admin já existe, pulando criação", p.info.Name)
		return nil
	}

	// Hash da senha padrão "admin123"
	hashedPassword, err := hashPassword("admin123")
	if err != nil {
		return fmt.Errorf("falha ao fazer hash da senha: %w", err)
	}

	// Criar usuário admin
	adminUser := User{
		ID:        uuid.New(),
		Username:  "admin",
		Email:     "admin@leonidas.com",
		Password:  hashedPassword,
		FirstName: "Administrador",
		LastName:  "do Sistema",
		IsActive:  true,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("falha ao criar usuário admin: %w", err)
	}

	// Associar role admin ao usuário
	var adminRole Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("falha ao encontrar role admin: %w", err)
	}

	userRole := UserRole{
		UserID:    adminUser.ID,
		RoleID:    adminRole.ID,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("falha ao associar role admin ao usuário: %w", err)
	}

	// Registrar no histórico
	history := UserHistory{
		ID:          uuid.New(),
		UserID:      adminUser.ID,
		Action:      "created",
		Description: "Usuário administrador criado automaticamente pelo sistema",
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&history).Error; err != nil {
		log.Printf("Aviso: Falha ao registrar histórico do usuário admin: %v", err)
	}

	log.Printf("[%s] Usuário administrador padrão criado com sucesso", p.info.Name)
	log.Printf("[%s] Username: admin, Password: admin123", p.info.Name)
	return nil
}

// hashPassword faz hash de uma senha
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
