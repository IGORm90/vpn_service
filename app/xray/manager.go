package xray

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"vpn-service/database"

	"github.com/xtls/xray-core/core"
)

// Manager управляет экземпляром Xray
type Manager struct {
	instance *core.Instance
	config   *Config
	mu       sync.RWMutex
	running  bool
}

// NewManager создает новый менеджер Xray
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Валидируем конфигурацию
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Создаем директории для логов
	logDir := filepath.Dir(config.AccessLogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{
		config:  config,
		running: false,
	}, nil
}

// Start запускает Xray сервер
func (m *Manager) Start(users []*database.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("xray is already running")
	}

	// Генерируем конфигурацию
	coreConfig, err := GenerateConfig(users, m.config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Создаем экземпляр Xray
	instance, err := core.New(coreConfig)
	if err != nil {
		return fmt.Errorf("failed to create xray instance: %w", err)
	}

	// Запускаем сервер
	if err := instance.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}

	m.instance = instance
	m.running = true

	log.Printf("Xray started successfully on port %d", m.config.Port)
	return nil
}

// Stop останавливает Xray сервер
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("xray is not running")
	}

	if m.instance != nil {
		if err := m.instance.Close(); err != nil {
			return fmt.Errorf("failed to stop xray: %w", err)
		}
	}

	m.instance = nil
	m.running = false

	log.Println("Xray stopped successfully")
	return nil
}

// Restart перезапускает Xray с новым списком пользователей
func (m *Manager) Restart(users []*database.User) error {
	log.Println("Restarting Xray...")

	// Останавливаем если запущен
	if m.IsRunning() {
		if err := m.Stop(); err != nil {
			return fmt.Errorf("failed to stop xray: %w", err)
		}
	}

	// Запускаем с новой конфигурацией
	if err := m.Start(users); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}

	log.Println("Xray restarted successfully")
	return nil
}

// IsRunning проверяет, запущен ли Xray
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetConfig возвращает текущую конфигурацию
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateUsers обновляет список пользователей (перезапускает сервер)
func (m *Manager) UpdateUsers(users []*database.User) error {
	log.Printf("Updating Xray users (total: %d, active: %d)",
		len(users), countActiveUsers(users))
	return m.Restart(users)
}

// AddUser добавляет пользователя (перезапускает сервер)
func (m *Manager) AddUser(users []*database.User) error {
	log.Printf("Adding user to Xray, total users: %d", len(users))
	return m.Restart(users)
}

// RemoveUser удаляет пользователя (перезапускает сервер)
func (m *Manager) RemoveUser(users []*database.User) error {
	log.Printf("Removing user from Xray, remaining users: %d", len(users))
	return m.Restart(users)
}

// countActiveUsers подсчитывает количество активных пользователей
func countActiveUsers(users []*database.User) int {
	count := 0
	for _, user := range users {
		if user.CanConnect() {
			count++
		}
	}
	return count
}
