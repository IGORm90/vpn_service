package controllers

import (
	"fmt"
	"net/http"
	"vpn-service/responses"
	"vpn-service/services"
)

// MainController обрабатывает системные HTTP запросы
type MainController struct {
	userService *services.UserService
}

// NewMainController создает новый экземпляр MainController
func NewMainController(userService *services.UserService) *MainController {
	return &MainController{
		userService: userService,
	}
}

// HealthCheck проверяет состояние сервиса
func (c *MainController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := c.userService.CheckHealth()
	responses.SendSuccess(w, status)
}

// GetStats возвращает статистику сервиса
func (c *MainController) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := c.userService.GetStats()
	if err != nil {
		responses.SendInternalError(w, fmt.Sprintf("Failed to get stats: %v", err))
		return
	}

	responses.SendSuccess(w, stats)
}
