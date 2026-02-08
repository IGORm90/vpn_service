package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vpn-service/responses"
	"vpn-service/services"

	"github.com/gorilla/mux"
)

// UserController обрабатывает HTTP запросы связанные с пользователями
type UserController struct {
	userService *services.UserService
}

// NewUserController создает новый экземпляр UserController
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// CreateUserRequest представляет запрос на создание пользователя
type CreateUserRequest struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// UpdateUserRequest представляет запрос на обновление пользователя
type UpdateUserRequest struct {
	IsActive *bool `json:"is_active,omitempty"`
}

// CreateUser создает нового пользователя
func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendBadRequest(w, "Invalid request body")
		return
	}

	dto := services.CreateUserDTO{
		ID:       req.ID,
		Username: req.Username,
	}

	user, err := c.userService.CreateUser(dto)
	if err != nil {
		switch err {
		case services.ErrInvalidUsername:
			responses.SendBadRequest(w, "Username is required")
		default:
			responses.SendInternalError(w, "Failed to create user")
		}
		return
	}

	responses.SendCreated(w, user)
}

// ListUsers возвращает список пользователей
func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Проверяем параметр запроса для фильтрации по активным пользователям
	activeOnly := r.URL.Query().Get("active") == "true"

	users, err := c.userService.ListUsers(activeOnly)
	if err != nil {
		responses.SendInternalError(w, "Failed to list users")
		return
	}

	responses.SendSuccess(w, users)
}

// GetUser возвращает пользователя по ID
func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responses.SendBadRequest(w, "Invalid user ID")
		return
	}

	user, err := c.userService.GetUser(id)
	if err != nil {
		if err == services.ErrUserNotFound {
			responses.SendNotFound(w, "User not found")
		} else {
			responses.SendInternalError(w, "Failed to get user")
		}
		return
	}

	responses.SendSuccess(w, user)
}

// UpdateUser обновляет данные пользователя
func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responses.SendBadRequest(w, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendBadRequest(w, "Invalid request body")
		return
	}

	dto := services.UpdateUserDTO{
		IsActive: req.IsActive,
	}

	user, err := c.userService.UpdateUser(id, dto)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			responses.SendNotFound(w, "User not found")
		default:
			responses.SendInternalError(w, "Failed to update user")
		}
		return
	}

	responses.SendSuccess(w, user)
}

// DeleteUser удаляет пользователя
func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responses.SendBadRequest(w, "Invalid user ID")
		return
	}

	if err := c.userService.DeleteUser(id); err != nil {
		if err == services.ErrUserNotFound {
			responses.SendNotFound(w, "User not found")
		} else {
			responses.SendInternalError(w, "Failed to delete user")
		}
		return
	}

	responses.SendSuccess(w, map[string]string{
		"message": "User deleted successfully",
	})
}

// GetUserConfig возвращает конфигурацию для подключения пользователя
func (c *UserController) GetUserConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responses.SendBadRequest(w, "Invalid user ID")
		return
	}

	config, err := c.userService.GetUserConfig(id)
	if err != nil {
		if err == services.ErrUserNotFound {
			responses.SendNotFound(w, "User not found")
		} else {
			responses.SendInternalError(w, "Failed to generate user config")
		}
		return
	}

	responses.SendSuccess(w, config)
}
