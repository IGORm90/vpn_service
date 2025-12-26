package api

import (
	"fmt"
	"net/http"
	"os"
	"vpn-service/database"
	"vpn-service/responses"
	"vpn-service/services"
	"vpn-service/xray"
)

// Handler содержит все зависимости для системных обработчиков
type Handler struct {
	userService *services.UserService
}

// NewHandler создает новый обработчик
func NewHandler(repo *database.Repository, xrayMgr *xray.Manager, xrayCfg *xray.Config) *Handler {
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		serverIP = "YOUR_SERVER_IP"
	}

	userService := services.NewUserService(repo, xrayMgr, xrayCfg, serverIP)

	return &Handler{
		userService: userService,
	}
}

// HealthCheck проверяет состояние сервиса
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := h.userService.CheckHealth()
	responses.SendSuccess(w, status)
}

// GetStats возвращает статистику сервиса
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userService.GetStats()
	if err != nil {
		responses.SendInternalError(w, fmt.Sprintf("Failed to get stats: %v", err))
		return
	}

	responses.SendSuccess(w, stats)
}
