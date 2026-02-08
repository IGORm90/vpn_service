package database

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository представляет репозиторий для работы с пользователями
type Repository struct {
	db *gorm.DB
}

var ErrDeviceLimitExceeded = errors.New("device limit exceeded")

// NewRepository создает новый экземпляр репозитория
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(id int64) (*User, error) {
	var user User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername возвращает пользователя по имени
func (r *Repository) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUUID возвращает пользователя по UUID
func (r *Repository) GetUserByUUID(uuid string) (*User, error) {
	var user User
	if err := r.db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// ListUsers возвращает список всех пользователей
func (r *Repository) ListUsers() ([]*User, error) {
	var users []*User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// ListActiveUsers возвращает список активных пользователей
func (r *Repository) ListActiveUsers() ([]*User, error) {
	var users []*User
	if err := r.db.Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}
	return users, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(user *User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeactivateUser деактивирует пользователя
func (r *Repository) DeactivateUser(id int64) error {
	result := r.db.Model(&User{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ActivateUser активирует пользователя
func (r *Repository) ActivateUser(id int64) error {
	result := r.db.Model(&User{}).
		Where("id = ?", id).
		Update("is_active", true)

	if result.Error != nil {
		return fmt.Errorf("failed to activate user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser удаляет пользователя
func (r *Repository) DeleteUser(id int64) error {
	result := r.db.Where("id = ?", id).Delete(&User{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// RecordUserDevice сохраняет информацию об устройстве пользователя по IP.
// Возвращает true если устройство новое.
func (r *Repository) RecordUserDevice(userID int64, ip string, maxDevices int) (bool, error) {
	if maxDevices <= 0 {
		return false, nil
	}
	if ip == "" || userID == 0 {
		return false, fmt.Errorf("invalid device params")
	}

	var existing UserDevice
	if err := r.db.Where("user_id = ? AND ip = ?", userID, ip).First(&existing).Error; err == nil {
		return false, r.db.Model(&UserDevice{}).
			Where("id = ?", existing.ID).
			Update("last_seen", time.Now()).Error
	} else if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to query device: %w", err)
	}

	var count int64
	if err := r.db.Model(&UserDevice{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to count devices: %w", err)
	}

	if int(count) >= maxDevices {
		return false, ErrDeviceLimitExceeded
	}

	device := &UserDevice{
		UserID:   userID,
		IP:       ip,
		LastSeen: time.Now(),
	}
	if err := r.db.Create(device).Error; err != nil {
		return false, fmt.Errorf("failed to create device: %w", err)
	}

	return true, nil
}

// CountUsers возвращает общее количество пользователей
func (r *Repository) CountUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// CountActiveUsers возвращает количество активных пользователей
func (r *Repository) CountActiveUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).
		Where("is_active = ?", true).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}
	return count, nil
}
