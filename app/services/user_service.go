package services

import (
	"errors"
	"fmt"
	"time"
	"vpn-service/database"
	"vpn-service/utils"
	"vpn-service/xray"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrInvalidUsername = errors.New("username is required")
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrCreateUser      = errors.New("failed to create user")
	ErrUpdateUser      = errors.New("failed to update user")
	ErrDeleteUser      = errors.New("failed to delete user")
	ErrListUsers       = errors.New("failed to list users")
	ErrGenerateConfig  = errors.New("failed to generate config")
)

// UserService содержит бизнес-логику для работы с пользователями
type UserService struct {
	repository  *database.Repository
	xrayManager *xray.Manager
	xrayConfig  *xray.Config
	serverIP    string
}

// NewUserService создает новый экземпляр UserService
func NewUserService(repo *database.Repository, xrayMgr *xray.Manager, xrayCfg *xray.Config, serverIP string) *UserService {
	return &UserService{
		repository:  repo,
		xrayManager: xrayMgr,
		xrayConfig:  xrayCfg,
		serverIP:    serverIP,
	}
}

// CreateUserDTO структура для создания пользователя
type CreateUserDTO struct {
	Username     string
	TrafficLimit int64
	ExpiresAt    time.Time
}

// UpdateUserDTO структура для обновления пользователя
type UpdateUserDTO struct {
	TrafficLimit *int64
	ExpiresAt    *time.Time
	IsActive     *bool
}

// UserConfigResponse структура ответа с конфигурацией пользователя
type UserConfigResponse struct {
	Username     string `json:"username"`
	UUID         string `json:"uuid"`
	ServerIP     string `json:"server_ip"`
	ServerPort   int    `json:"server_port"`
	JSON         string `json:"json"`
	URI          string `json:"uri"`
	QRCode       string `json:"qr_code"`
	ExpiresAt    string `json:"expires_at"`
	TrafficLimit int64  `json:"traffic_limit"`
	TrafficUsed  int64  `json:"traffic_used"`
	IsActive     bool   `json:"is_active"`
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(dto CreateUserDTO) (*database.User, error) {
	// Валидация
	if dto.Username == "" {
		return nil, ErrInvalidUsername
	}

	// Проверяем уникальность
	if _, err := s.repository.GetUserByUsername(dto.Username); err == nil {
		return nil, ErrUsernameExists
	}

	// Создаем пользователя
	user := &database.User{
		Username:     dto.Username,
		UUID:         utils.GenerateUUID(),
		IsActive:     true,
		TrafficLimit: dto.TrafficLimit,
		ExpiresAt:    dto.ExpiresAt,
	}

	if err := s.repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateUser, err)
	}

	// Обновляем Xray конфигурацию
	if err := s.syncXrayUsers(); err != nil {
		// Логируем ошибку, но не возвращаем её
		// так как пользователь уже создан в БД
		fmt.Printf("Warning: failed to sync Xray users: %v\n", err)
	}

	return user, nil
}

// ListUsers возвращает список пользователей
func (s *UserService) ListUsers(activeOnly bool) ([]*database.User, error) {
	var users []*database.User
	var err error

	if activeOnly {
		users, err = s.repository.ListActiveUsers()
	} else {
		users, err = s.repository.ListUsers()
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListUsers, err)
	}

	return users, nil
}

// GetUser возвращает пользователя по ID
func (s *UserService) GetUser(id uint) (*database.User, error) {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(id uint, dto UpdateUserDTO) (*database.User, error) {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Обновляем поля если они указаны
	if dto.TrafficLimit != nil {
		user.TrafficLimit = *dto.TrafficLimit
	}

	if dto.ExpiresAt != nil {
		user.ExpiresAt = *dto.ExpiresAt
	}

	if dto.IsActive != nil {
		user.IsActive = *dto.IsActive
	}

	if err := s.repository.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpdateUser, err)
	}

	// Обновляем Xray
	if err := s.syncXrayUsers(); err != nil {
		fmt.Printf("Warning: failed to sync Xray users: %v\n", err)
	}

	return user, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(id uint) error {
	if err := s.repository.DeleteUser(id); err != nil {
		return ErrUserNotFound
	}

	// Обновляем Xray
	if err := s.syncXrayUsers(); err != nil {
		fmt.Printf("Warning: failed to sync Xray users: %v\n", err)
	}

	return nil
}

// GetUserConfig возвращает конфигурацию для подключения пользователя
func (s *UserService) GetUserConfig(id uint) (*UserConfigResponse, error) {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Генерируем конфигурации
	jsonConfig, err := xray.GenerateClientJSON(user, s.xrayConfig, s.serverIP)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate JSON config: %v", ErrGenerateConfig, err)
	}

	vlessURI, err := xray.GenerateVlessURI(user, s.xrayConfig, s.serverIP)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate VLESS URI: %v", ErrGenerateConfig, err)
	}

	qrCode, err := utils.GenerateQRCode(vlessURI)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate QR code: %v", ErrGenerateConfig, err)
	}

	response := &UserConfigResponse{
		Username:     user.Username,
		UUID:         user.UUID,
		ServerIP:     s.serverIP,
		ServerPort:   s.xrayConfig.Port,
		JSON:         jsonConfig,
		URI:          vlessURI,
		QRCode:       qrCode,
		ExpiresAt:    user.ExpiresAt.Format(time.RFC3339),
		TrafficLimit: user.TrafficLimit,
		TrafficUsed:  user.TrafficUsed,
		IsActive:     user.IsActive,
	}

	return response, nil
}

// ResetUserTraffic сбрасывает счетчик трафика пользователя
func (s *UserService) ResetUserTraffic(id uint) error {
	if err := s.repository.ResetTraffic(id); err != nil {
		return ErrUserNotFound
	}
	return nil
}

// GetStats возвращает статистику по пользователям
func (s *UserService) GetStats() (map[string]interface{}, error) {
	totalUsers, _ := s.repository.CountUsers()
	activeUsers, _ := s.repository.CountActiveUsers()
	expiredUsers, _ := s.repository.CountExpiredUsers()
	overLimitUsers, _ := s.repository.CountUsersOverLimit()

	stats := map[string]interface{}{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"expired_users":    expiredUsers,
		"over_limit_users": overLimitUsers,
		"xray_running":     s.xrayManager.IsRunning(),
	}

	return stats, nil
}

// CheckHealth проверяет состояние сервиса
func (s *UserService) CheckHealth() map[string]interface{} {
	status := map[string]interface{}{
		"status":      "healthy",
		"time":        time.Now().Format(time.RFC3339),
		"xray_status": s.xrayManager.IsRunning(),
	}

	// Проверяем БД
	if _, err := s.repository.CountUsers(); err != nil {
		status["database"] = "error"
		status["status"] = "degraded"
	} else {
		status["database"] = "ok"
	}

	return status
}

// syncXrayUsers синхронизирует пользователей с Xray
func (s *UserService) syncXrayUsers() error {
	users, err := s.repository.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}

	if err := s.xrayManager.UpdateUsers(users); err != nil {
		return fmt.Errorf("failed to update Xray: %v", err)
	}

	return nil
}
