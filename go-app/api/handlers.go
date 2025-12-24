package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"vpn-service/database"
	"vpn-service/utils"
	"vpn-service/xray"

	"github.com/gorilla/mux"
)

// Handler содержит все зависимости для обработчиков
type Handler struct {
	repository  *database.Repository
	xrayManager *xray.Manager
	xrayConfig  *xray.Config
	serverIP    string
}

// NewHandler создает новый обработчик
func NewHandler(repo *database.Repository, xrayMgr *xray.Manager, xrayCfg *xray.Config) *Handler {
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		serverIP = "YOUR_SERVER_IP"
	}

	return &Handler{
		repository:  repo,
		xrayManager: xrayMgr,
		xrayConfig:  xrayCfg,
		serverIP:    serverIP,
	}
}

// CreateUserRequest представляет запрос на создание пользователя
type CreateUserRequest struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	TrafficLimit int64     `json:"traffic_limit,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

// UpdateUserRequest представляет запрос на обновление пользователя
type UpdateUserRequest struct {
	Password     *string    `json:"password,omitempty"`
	TrafficLimit *int64     `json:"traffic_limit,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// CreateUser создает нового пользователя
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendBadRequest(w, "Invalid request body")
		return
	}

	// Валидация
	if req.Username == "" {
		SendBadRequest(w, "Username is required")
		return
	}
	if req.Password == "" {
		SendBadRequest(w, "Password is required")
		return
	}

	// Проверяем уникальность
	if _, err := h.repository.GetUserByUsername(req.Username); err == nil {
		SendBadRequest(w, "Username already exists")
		return
	}

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		SendInternalError(w, "Failed to hash password")
		return
	}

	// Создаем пользователя
	user := &database.User{
		Username:     req.Username,
		Password:     hashedPassword,
		UUID:         utils.GenerateUUID(),
		IsActive:     true,
		TrafficLimit: req.TrafficLimit,
		ExpiresAt:    req.ExpiresAt,
	}

	if err := h.repository.CreateUser(user); err != nil {
		SendInternalError(w, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	// Обновляем Xray конфигурацию
	users, _ := h.repository.ListUsers()
	if err := h.xrayManager.UpdateUsers(users); err != nil {
		// Логируем ошибку, но не возвращаем её клиенту
		// так как пользователь уже создан в БД
		fmt.Printf("Warning: failed to update Xray: %v\n", err)
	}

	SendCreated(w, user)
}

// ListUsers возвращает список всех пользователей
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Парсим query параметры
	activeOnly := r.URL.Query().Get("active") == "true"

	var users []*database.User
	var err error

	if activeOnly {
		users, err = h.repository.ListActiveUsers()
	} else {
		users, err = h.repository.ListUsers()
	}

	if err != nil {
		SendInternalError(w, fmt.Sprintf("Failed to list users: %v", err))
		return
	}

	SendSuccess(w, users)
}

// GetUser возвращает информацию о пользователе
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendBadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.repository.GetUserByID(uint(id))
	if err != nil {
		SendNotFound(w, "User not found")
		return
	}

	SendSuccess(w, user)
}

// UpdateUser обновляет данные пользователя
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendBadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.repository.GetUserByID(uint(id))
	if err != nil {
		SendNotFound(w, "User not found")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendBadRequest(w, "Invalid request body")
		return
	}

	// Обновляем поля если они указаны
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			SendInternalError(w, "Failed to hash password")
			return
		}
		user.Password = hashedPassword
	}

	if req.TrafficLimit != nil {
		user.TrafficLimit = *req.TrafficLimit
	}

	if req.ExpiresAt != nil {
		user.ExpiresAt = *req.ExpiresAt
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.repository.UpdateUser(user); err != nil {
		SendInternalError(w, fmt.Sprintf("Failed to update user: %v", err))
		return
	}

	// Обновляем Xray
	users, _ := h.repository.ListUsers()
	h.xrayManager.UpdateUsers(users)

	SendSuccess(w, user)
}

// DeleteUser удаляет пользователя
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendBadRequest(w, "Invalid user ID")
		return
	}

	if err := h.repository.DeleteUser(uint(id)); err != nil {
		SendNotFound(w, "User not found")
		return
	}

	// Обновляем Xray
	users, _ := h.repository.ListUsers()
	h.xrayManager.UpdateUsers(users)

	SendNoContent(w)
}

// GetUserConfig возвращает конфигурацию клиента для пользователя
func (h *Handler) GetUserConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendBadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.repository.GetUserByID(uint(id))
	if err != nil {
		SendNotFound(w, "User not found")
		return
	}

	// Генерируем конфигурации
	jsonConfig, _ := xray.GenerateClientJSON(user, h.xrayConfig, h.serverIP)
	vlessURI, _ := xray.GenerateVlessURI(user, h.xrayConfig, h.serverIP)
	qrCode, _ := utils.GenerateQRCode(vlessURI)

	response := xray.ClientConfigResponse{
		Username:     user.Username,
		UUID:         user.UUID,
		ServerIP:     h.serverIP,
		ServerPort:   h.xrayConfig.Port,
		JSON:         jsonConfig,
		URI:          vlessURI,
		QRCode:       qrCode,
		ExpiresAt:    user.ExpiresAt.Format(time.RFC3339),
		TrafficLimit: user.TrafficLimit,
		TrafficUsed:  user.TrafficUsed,
		IsActive:     user.IsActive,
	}

	SendSuccess(w, response)
}

// ResetTraffic сбрасывает счетчик трафика пользователя
func (h *Handler) ResetTraffic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		SendBadRequest(w, "Invalid user ID")
		return
	}

	if err := h.repository.ResetTraffic(uint(id)); err != nil {
		SendNotFound(w, "User not found")
		return
	}

	SendSuccess(w, map[string]string{
		"message": "Traffic reset successfully",
	})
}

// HealthCheck проверяет состояние сервиса
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":      "healthy",
		"time":        time.Now().Format(time.RFC3339),
		"xray_status": h.xrayManager.IsRunning(),
	}

	// Проверяем БД
	if _, err := h.repository.CountUsers(); err != nil {
		status["database"] = "error"
		status["status"] = "degraded"
	} else {
		status["database"] = "ok"
	}

	SendSuccess(w, status)
}

// GetStats возвращает статистику сервиса
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	totalUsers, _ := h.repository.CountUsers()
	activeUsers, _ := h.repository.CountActiveUsers()
	expiredUsers, _ := h.repository.CountExpiredUsers()
	overLimitUsers, _ := h.repository.CountUsersOverLimit()

	stats := map[string]interface{}{
		"total_users":      totalUsers,
		"active_users":     activeUsers,
		"expired_users":    expiredUsers,
		"over_limit_users": overLimitUsers,
		"xray_running":     h.xrayManager.IsRunning(),
	}

	SendSuccess(w, stats)
}
