package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"leonidas/core/pkg/plugin"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTService gerencia tokens JWT
type JWTService struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshDuration time.Duration
}

// NewJWTService cria uma nova instância do serviço JWT
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:       secretKey,
		tokenDuration:   15 * time.Minute,   // Token expira em 15 minutos
		refreshDuration: 7 * 24 * time.Hour, // Refresh token expira em 7 dias
	}
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse representa a resposta de login
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	User         User   `json:"user"`
}

// RefreshTokenRequest representa a requisição de refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ChangePasswordRequest representa a requisição de troca de senha
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ForgotPasswordRequest representa a requisição de recuperação de senha
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest representa a requisição de reset de senha
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// AuthService gerencia autenticação e autorização
type AuthService struct {
	userRepo          *UserRepository
	sessionRepo       *UserSessionRepository
	passwordResetRepo *PasswordResetRepository
	historyRepo       *UserHistoryRepository
	jwtService        *JWTService
	cacheService      *CacheService
}

// NewAuthService cria uma nova instância do serviço de autenticação
func NewAuthService(
	userRepo *UserRepository,
	sessionRepo *UserSessionRepository,
	passwordResetRepo *PasswordResetRepository,
	historyRepo *UserHistoryRepository,
	jwtService *JWTService,
	cacheService *CacheService,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: passwordResetRepo,
		historyRepo:       historyRepo,
		jwtService:        jwtService,
		cacheService:      cacheService,
	}
}

// Login autentica um usuário
func (s *AuthService) Login(ctx context.Context, req LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// Verificar se o repositório está disponível
	if s.userRepo == nil {
		return nil, fmt.Errorf("repositório de usuários não disponível")
	}

	// Buscar usuário por username ou email
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		// Tentar buscar por email
		user, err = s.userRepo.GetByEmail(req.Username)
		if err != nil {
			return nil, fmt.Errorf("credenciais inválidas")
		}
	}

	// Verificar se o usuário está ativo e habilitado
	if !user.IsActive || !user.IsEnabled {
		return nil, fmt.Errorf("usuário desabilitado")
	}

	// Verificar senha
	if !s.checkPassword(req.Password, user.Password) {
		// Registrar tentativa de login falhada
		s.recordLoginAttempt(ctx, user.ID, false, ipAddress, userAgent)
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// Gerar tokens
	token, refreshToken, expiresAt, err := s.generateTokens(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar tokens: %w", err)
	}

	// Criar sessão
	session := &UserSession{
		ID:               uuid.New(),
		UserID:           user.ID,
		Token:            token,
		RefreshToken:     refreshToken,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: time.Now().Add(s.jwtService.refreshDuration),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		log.Printf("Aviso: Falha ao criar sessão: %v", err)
	}

	// Atualizar último login
	if err := s.userRepo.UpdateLastLogin(user.ID.String()); err != nil {
		log.Printf("Aviso: Falha ao atualizar último login: %v", err)
	}

	// Registrar login bem-sucedido
	s.recordLoginAttempt(ctx, user.ID, true, ipAddress, userAgent)

	// Invalidar cache do usuário
	s.cacheService.InvalidateUserCache(ctx, user.ID.String())

	// Remover senha da resposta
	user.Password = ""

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User:         *user,
	}, nil
}

// RefreshToken renova um token JWT
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	// Buscar sessão pelo refresh token
	session, err := s.sessionRepo.GetByRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token inválido")
	}

	// Verificar se a sessão ainda é válida
	if !session.IsActive || session.IsRefreshExpired() {
		return nil, fmt.Errorf("refresh token expirado")
	}

	// Buscar usuário
	user, err := s.userRepo.GetByID(session.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	// Verificar se o usuário ainda está ativo
	if !user.IsActive || !user.IsEnabled {
		return nil, fmt.Errorf("usuário desabilitado")
	}

	// Gerar novos tokens
	newToken, newRefreshToken, expiresAt, err := s.generateTokens(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar novos tokens: %w", err)
	}

	// Atualizar sessão
	session.Token = newToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = expiresAt
	session.RefreshExpiresAt = time.Now().Add(s.jwtService.refreshDuration)
	session.UpdatedAt = time.Now()

	if err := s.sessionRepo.Create(session); err != nil {
		log.Printf("Aviso: Falha ao atualizar sessão: %v", err)
	}

	// Remover senha da resposta
	user.Password = ""

	return &LoginResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User:         *user,
	}, nil
}

// Logout faz logout de um usuário
func (s *AuthService) Logout(ctx context.Context, token string) error {
	// Desativar sessão
	if err := s.sessionRepo.Deactivate(token); err != nil {
		log.Printf("Aviso: Falha ao desativar sessão: %v", err)
	}

	// Invalidar cache
	// TODO: Implementar quando tivermos o token no contexto
	return nil
}

// ChangePassword altera a senha de um usuário
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	// Buscar usuário
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("usuário não encontrado")
	}

	// Verificar senha atual
	if !s.checkPassword(req.CurrentPassword, user.Password) {
		return fmt.Errorf("senha atual incorreta")
	}

	// Hash da nova senha
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("falha ao processar nova senha: %w", err)
	}

	// Atualizar senha
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("falha ao atualizar senha: %w", err)
	}

	// Registrar no histórico
	history := &UserHistory{
		ID:          uuid.New(),
		UserID:      user.ID,
		Action:      "password_changed",
		Description: "Senha alterada pelo usuário",
		CreatedAt:   time.Now(),
	}
	s.historyRepo.Create(history)

	// Invalidar todas as sessões do usuário (força novo login)
	s.sessionRepo.DeactivateAllUserSessions(userID)

	// Invalidar cache
	s.cacheService.InvalidateUserCache(ctx, userID)

	log.Printf("Senha alterada para usuário %s", user.Username)
	return nil
}

// ForgotPassword inicia o processo de recuperação de senha
func (s *AuthService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	// Buscar usuário por email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Por segurança, não revelar se o email existe ou não
		log.Printf("Tentativa de recuperação de senha para email inexistente: %s", req.Email)
		return nil
	}

	// Verificar se o usuário está ativo
	if !user.IsActive || !user.IsEnabled {
		log.Printf("Tentativa de recuperação de senha para usuário desabilitado: %s", user.Email)
		return nil
	}

	// Gerar token de reset
	token, err := s.generateResetToken()
	if err != nil {
		return fmt.Errorf("falha ao gerar token de reset: %w", err)
	}

	// Criar registro de reset
	reset := &PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token expira em 24 horas
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := s.passwordResetRepo.Create(reset); err != nil {
		return fmt.Errorf("falha ao criar token de reset: %w", err)
	}

	// Registrar no histórico
	history := &UserHistory{
		ID:          uuid.New(),
		UserID:      user.ID,
		Action:      "password_reset_requested",
		Description: "Solicitação de recuperação de senha",
		CreatedAt:   time.Now(),
	}
	s.historyRepo.Create(history)

	// TODO: Enviar email com o token
	log.Printf("Token de reset gerado para usuário %s: %s", user.Email, token)

	return nil
}

// ResetPassword redefine a senha usando um token
func (s *AuthService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	// Buscar token de reset
	reset, err := s.passwordResetRepo.GetByToken(req.Token)
	if err != nil {
		return fmt.Errorf("token de reset inválido")
	}

	// Verificar se o token não expirou
	if reset.IsExpired() {
		return fmt.Errorf("token de reset expirado")
	}

	// Buscar usuário
	user, err := s.userRepo.GetByID(reset.UserID.String())
	if err != nil {
		return fmt.Errorf("usuário não encontrado")
	}

	// Verificar se o usuário ainda está ativo
	if !user.IsActive || !user.IsEnabled {
		return fmt.Errorf("usuário desabilitado")
	}

	// Hash da nova senha
	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("falha ao processar nova senha: %w", err)
	}

	// Atualizar senha
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("falha ao atualizar senha: %w", err)
	}

	// Marcar token como usado
	s.passwordResetRepo.MarkAsUsed(req.Token)

	// Invalidar todas as sessões do usuário
	s.sessionRepo.DeactivateAllUserSessions(user.ID.String())

	// Registrar no histórico
	history := &UserHistory{
		ID:          uuid.New(),
		UserID:      user.ID,
		Action:      "password_reset_completed",
		Description: "Senha redefinida via token de reset",
		CreatedAt:   time.Now(),
	}
	s.historyRepo.Create(history)

	// Invalidar cache
	s.cacheService.InvalidateUserCache(ctx, user.ID.String())

	log.Printf("Senha redefinida para usuário %s via token de reset", user.Username)
	return nil
}

// ValidateToken valida um token JWT
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*User, error) {
	// Buscar sessão
	session, err := s.sessionRepo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("token inválido")
	}

	// Verificar se a sessão ainda é válida
	if !session.IsActive || session.IsExpired() {
		return nil, fmt.Errorf("token expirado")
	}

	// Buscar usuário
	user, err := s.userRepo.GetByID(session.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	// Verificar se o usuário ainda está ativo
	if !user.IsActive || !user.IsEnabled {
		return nil, fmt.Errorf("usuário desabilitado")
	}

	return user, nil
}

// Helper functions

// hashPassword faz hash de uma senha
func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword verifica se uma senha está correta
func (s *AuthService) checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateTokens gera tokens JWT e refresh token
func (s *AuthService) generateTokens(userID string) (string, string, time.Time, error) {
	// Por enquanto, vamos gerar tokens simples
	// TODO: Implementar JWT real quando adicionarmos a dependência

	token := s.generateRandomToken(32)
	refreshToken := s.generateRandomToken(32)
	expiresAt := time.Now().Add(s.jwtService.tokenDuration)

	return token, refreshToken, expiresAt, nil
}

// generateResetToken gera um token de reset
func (s *AuthService) generateResetToken() (string, error) {
	return s.generateRandomToken(32), nil
}

// generateRandomToken gera um token aleatório
func (s *AuthService) generateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// recordLoginAttempt registra uma tentativa de login
func (s *AuthService) recordLoginAttempt(ctx context.Context, userID uuid.UUID, success bool, ipAddress, userAgent string) {
	action := "login_failed"
	description := "Tentativa de login falhada"

	if success {
		action = "login_success"
		description = "Login realizado com sucesso"
	}

	history := &UserHistory{
		ID:          uuid.New(),
		UserID:      userID,
		Action:      action,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	if err := s.historyRepo.Create(history); err != nil {
		log.Printf("Aviso: Falha ao registrar tentativa de login: %v", err)
	}
}

// extractTokenFromHeader extrai o token do header Authorization
func extractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// getClientInfo extrai informações do cliente da requisição
func getClientInfo(request plugin.HTTPRequest) (ipAddress, userAgent string) {
	ipAddress = request.Headers["X-Forwarded-For"]
	if ipAddress == "" {
		ipAddress = request.Headers["X-Real-IP"]
	}
	if ipAddress == "" {
		ipAddress = "unknown"
	}

	userAgent = request.Headers["User-Agent"]
	if userAgent == "" {
		userAgent = "unknown"
	}

	return ipAddress, userAgent
}
