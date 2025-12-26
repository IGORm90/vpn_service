package monitoring

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"vpn-service/database"

	"github.com/nxadm/tail"
)

// LogEntry представляет запись в логе Xray
type LogEntry struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	Email    string `json:"email"`
	UUID     string `json:"uuid"`
	Source   string `json:"source"`
	Dest     string `json:"dest"`
	Protocol string `json:"protocol"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// TrafficStats хранит статистику трафика для пользователя
type TrafficStats struct {
	UUID     string
	Email    string
	Upload   int64
	Download int64
	LastSeen time.Time
	mu       sync.Mutex
}

// LogMonitor мониторит логи Xray и обновляет статистику
type LogMonitor struct {
	logPath    string
	repository *database.Repository
	stats      map[string]*TrafficStats
	mu         sync.RWMutex
	interval   time.Duration
	stopCh     chan struct{}
	running    bool
}

// NewLogMonitor создает новый монитор логов
func NewLogMonitor(logPath string, repo *database.Repository, updateInterval time.Duration) *LogMonitor {
	return &LogMonitor{
		logPath:    logPath,
		repository: repo,
		stats:      make(map[string]*TrafficStats),
		interval:   updateInterval,
		stopCh:     make(chan struct{}),
		running:    false,
	}
}

// Start запускает мониторинг логов
func (m *LogMonitor) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("log monitor is already running")
	}
	m.running = true
	m.mu.Unlock()

	// Проверяем существование файла лога
	if _, err := os.Stat(m.logPath); os.IsNotExist(err) {
		log.Printf("Log file does not exist yet, creating: %s", m.logPath)
		if err := os.MkdirAll(filepath.Dir(m.logPath), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		if _, err := os.Create(m.logPath); err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
	}

	// Запускаем tail логов в отдельной горутине
	go m.tailLogs()

	// Запускаем периодическое обновление БД
	go m.periodicUpdate()

	log.Printf("Log monitor started, watching: %s", m.logPath)
	return nil
}

// Stop останавливает мониторинг
func (m *LogMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	close(m.stopCh)
	m.running = false

	// Финальное обновление статистики в БД
	m.flushStats()

	log.Println("Log monitor stopped")
}

// tailLogs читает логи в реальном времени
func (m *LogMonitor) tailLogs() {
	// Используем библиотеку tail для чтения логов
	t, err := tail.TailFile(m.logPath, tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: os.SEEK_END,
		},
	})
	if err != nil {
		log.Printf("Failed to tail log file: %v", err)
		return
	}

	for {
		select {
		case line := <-t.Lines:
			if line == nil {
				continue
			}
			m.processLogLine(line.Text)
		case <-m.stopCh:
			t.Stop()
			return
		}
	}
}

// processLogLine обрабатывает строку лога
func (m *LogMonitor) processLogLine(line string) {
	// Пытаемся распарсить как JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		// Если не JSON, пытаемся распарсить как обычный текст
		m.parseTextLog(line)
		return
	}

	// Обновляем статистику
	m.updateStats(entry.Email, entry.UUID, entry.Upload, entry.Download)
}

// parseTextLog парсит текстовый лог (fallback)
func (m *LogMonitor) parseTextLog(line string) {
	// Xray логи могут быть в формате:
	// 2024/12/24 12:00:00 [Info] [email: user@example.com] accepted connection
	// Простой парсинг для извлечения email
	if strings.Contains(line, "accepted") || strings.Contains(line, "connection") {
		// Извлекаем email если есть
		if idx := strings.Index(line, "email:"); idx != -1 {
			rest := line[idx+6:]
			endIdx := strings.IndexAny(rest, " ]")
			if endIdx != -1 {
				email := strings.TrimSpace(rest[:endIdx])
				// Обновляем только lastSeen для этого пользователя
				m.updateLastSeen(email)
			}
		}
	}
}

// updateStats обновляет статистику трафика
func (m *LogMonitor) updateStats(email, uuid string, upload, download int64) {
	if email == "" && uuid == "" {
		return
	}

	key := uuid
	if key == "" {
		key = email
	}

	m.mu.Lock()
	stat, exists := m.stats[key]
	if !exists {
		stat = &TrafficStats{
			UUID:  uuid,
			Email: email,
		}
		m.stats[key] = stat
	}
	m.mu.Unlock()

	stat.mu.Lock()
	stat.Upload += upload
	stat.Download += download
	stat.LastSeen = time.Now()
	stat.mu.Unlock()
}

// updateLastSeen обновляет время последнего подключения
func (m *LogMonitor) updateLastSeen(email string) {
	m.mu.Lock()
	stat, exists := m.stats[email]
	if !exists {
		stat = &TrafficStats{
			Email: email,
		}
		m.stats[email] = stat
	}
	m.mu.Unlock()

	stat.mu.Lock()
	stat.LastSeen = time.Now()
	stat.mu.Unlock()
}

// periodicUpdate периодически сохраняет статистику в БД
func (m *LogMonitor) periodicUpdate() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.flushStats()
		case <-m.stopCh:
			return
		}
	}
}

// flushStats сохраняет накопленную статистику в БД
func (m *LogMonitor) flushStats() {
	m.mu.RLock()
	// Копируем статистику для обработки
	statsCopy := make(map[string]*TrafficStats)
	for k, v := range m.stats {
		statsCopy[k] = v
	}
	m.mu.RUnlock()

	for key, stat := range statsCopy {
		stat.mu.Lock()
		upload := stat.Upload
		download := stat.Download
		uuid := stat.UUID
		email := stat.Email
		stat.mu.Unlock()

		if upload == 0 && download == 0 {
			continue
		}

		// Находим пользователя по UUID или email
		var user *database.User
		var err error

		if uuid != "" {
			user, err = m.repository.GetUserByUUID(uuid)
		} else if email != "" {
			user, err = m.repository.GetUserByUsername(email)
		}

		if err != nil {
			log.Printf("Failed to find user %s: %v", key, err)
			continue
		}

		// Обновляем трафик в БД
		if err := m.repository.UpdateTrafficUsage(user.UUID, upload, download); err != nil {
			log.Printf("Failed to update traffic for user %s: %v", user.Username, err)
			continue
		}

		log.Printf("Updated traffic for user %s: +%d up, +%d down (total: %d)",
			user.Username, upload, download, user.TrafficUsed+upload+download)

		// Сбрасываем локальные счетчики
		stat.mu.Lock()
		stat.Upload = 0
		stat.Download = 0
		stat.mu.Unlock()
	}
}

// GetStats возвращает текущую статистику
func (m *LogMonitor) GetStats() map[string]*TrafficStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Возвращаем копию
	statsCopy := make(map[string]*TrafficStats)
	for k, v := range m.stats {
		v.mu.Lock()
		statsCopy[k] = &TrafficStats{
			UUID:     v.UUID,
			Email:    v.Email,
			Upload:   v.Upload,
			Download: v.Download,
			LastSeen: v.LastSeen,
		}
		v.mu.Unlock()
	}

	return statsCopy
}

// IsRunning проверяет, запущен ли монитор
func (m *LogMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// SimpleLogMonitor для случаев когда нет хвостового чтения
type SimpleLogMonitor struct {
	logPath    string
	repository *database.Repository
	lastPos    int64
	stopCh     chan struct{}
}

// NewSimpleLogMonitor создает простой монитор который читает файл периодически
func NewSimpleLogMonitor(logPath string, repo *database.Repository) *SimpleLogMonitor {
	return &SimpleLogMonitor{
		logPath:    logPath,
		repository: repo,
		lastPos:    0,
		stopCh:     make(chan struct{}),
	}
}

// Start запускает простой мониторинг
func (s *SimpleLogMonitor) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkLogs()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop останавливает мониторинг
func (s *SimpleLogMonitor) Stop() {
	close(s.stopCh)
}

// checkLogs проверяет новые записи в логе
func (s *SimpleLogMonitor) checkLogs() {
	file, err := os.Open(s.logPath)
	if err != nil {
		return
	}
	defer file.Close()

	// Перемещаемся на последнюю позицию
	if _, err := file.Seek(s.lastPos, 0); err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Простой парсинг строки
		_ = line // TODO: обработка
	}

	// Сохраняем текущую позицию
	pos, _ := file.Seek(0, os.SEEK_CUR)
	s.lastPos = pos
}
