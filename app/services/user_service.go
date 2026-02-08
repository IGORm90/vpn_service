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
	ID       int64
	Username string
}

// UpdateUserDTO структура для обновления пользователя
type UpdateUserDTO struct {
	IsActive *bool
}

// UserConfigResponse структура ответа с конфигурацией пользователя
type UserConfigResponse struct {
	Username   string `json:"username"`
	UUID       string `json:"uuid"`
	ServerIP   string `json:"server_ip"`
	ServerPort int    `json:"server_port"`
	JSON       string `json:"json"`
	URI        string `json:"uri"`
	QRCode     string `json:"qr_code"`
	IsActive   bool   `json:"is_active"`
}

// CreateUser создает нового пользователя или возвращает существующего
func (s *UserService) CreateUser(dto CreateUserDTO) (*database.User, error) {
	// Валидация
	if dto.Username == "" {
		return nil, ErrInvalidUsername
	}
	if dto.ID == 0 {
		return nil, ErrInvalidUserID
	}

	// Проверяем существование пользователя
	if existingUser, err := s.repository.GetUserByUsername(dto.Username); err == nil {
		// Пользователь уже существует - возвращаем его
		return existingUser, nil
	}

	// Создаем нового пользователя
	user := &database.User{
		ID:       dto.ID,
		Username: dto.Username,
		UUID:     utils.GenerateUUID(),
		IsActive: true,
	}

	if err := s.repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateUser, err)
	}

	// Hot-update Xray (fallback to full restart on error)
	if user.CanConnect() {
		s.hotAddUserWithFallback(user)
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
func (s *UserService) GetUser(id int64) (*database.User, error) {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(id int64, dto UpdateUserDTO) (*database.User, error) {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	oldCanConnect := user.CanConnect()

	// Обновляем поля если они указаны
	if dto.IsActive != nil {
		user.IsActive = *dto.IsActive
	}

	if err := s.repository.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpdateUser, err)
	}

	s.hotUpdateUserAccess(user, oldCanConnect)

	return user, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(id int64) error {
	user, err := s.repository.GetUserByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.repository.DeleteUser(id); err != nil {
		return ErrUserNotFound
	}

	if user.CanConnect() {
		s.hotRemoveUserWithFallback(user)
	}

	return nil
}

// GetUserConfig возвращает конфигурацию для подключения пользователя
func (s *UserService) GetUserConfig(id int64) (*UserConfigResponse, error) {
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
		Username:   user.Username,
		UUID:       user.UUID,
		ServerIP:   s.serverIP,
		ServerPort: s.xrayConfig.Port,
		JSON:       jsonConfig,
		URI:        vlessURI,
		QRCode:     qrCode,
		IsActive:   user.IsActive,
	}

	return response, nil
}

// GetStats возвращает статистику по пользователям
func (s *UserService) GetStats() (map[string]interface{}, error) {
	totalUsers, _ := s.repository.CountUsers()
	activeUsers, _ := s.repository.CountActiveUsers()

	stats := map[string]interface{}{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"xray_running": s.xrayManager.IsRunning(),
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

func (s *UserService) hotAddUserWithFallback(user *database.User) {
	if err := s.xrayManager.AddUserHot(user); err != nil {
		s.fallbackXraySync("hot-add user", err)
	}
}

func (s *UserService) hotRemoveUserWithFallback(user *database.User) {
	if err := s.xrayManager.RemoveUserHot(user); err != nil {
		s.fallbackXraySync("hot-remove user", err)
	}
}

func (s *UserService) hotUpdateUserAccess(user *database.User, oldCanConnect bool) {
	newCanConnect := user.CanConnect()
	if oldCanConnect == newCanConnect {
		return
	}

	if newCanConnect {
		s.hotAddUserWithFallback(user)
		return
	}
	s.hotRemoveUserWithFallback(user)
}

func (s *UserService) fallbackXraySync(action string, err error) {
	fmt.Printf("Warning: failed to %s: %v\n", action, err)
	if err := s.syncXrayUsers(); err != nil {
		fmt.Printf("Warning: failed to sync Xray users: %v\n", err)
	}
}
