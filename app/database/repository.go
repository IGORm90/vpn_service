package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository представляет репозиторий для работы с пользователями
type Repository struct {
	db *gorm.DB
}

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
func (r *Repository) GetUserByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
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
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
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

// UpdateTrafficUsage обновляет использованный трафик пользователя
func (r *Repository) UpdateTrafficUsage(uuid string, upload, download int64) error {
	totalTraffic := upload + download

	result := r.db.Model(&User{}).
		Where("uuid = ?", uuid).
		UpdateColumn("traffic_used", gorm.Expr("traffic_used + ?", totalTraffic))

	if result.Error != nil {
		return fmt.Errorf("failed to update traffic usage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Проверяем, не превышен ли лимит
	user, err := r.GetUserByUUID(uuid)
	if err != nil {
		return err
	}

	// Автоматически деактивируем пользователя если превышен лимит
	if user.IsOverLimit() && user.IsActive {
		user.IsActive = false
		if err := r.UpdateUser(user); err != nil {
			return fmt.Errorf("failed to deactivate user over limit: %w", err)
		}
	}

	return nil
}

// DeactivateUser деактивирует пользователя
func (r *Repository) DeactivateUser(id uint) error {
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
func (r *Repository) ActivateUser(id uint) error {
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
func (r *Repository) DeleteUser(id uint) error {
	result := r.db.Delete(&User{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ResetTraffic сбрасывает счетчик трафика пользователя
func (r *Repository) ResetTraffic(id uint) error {
	result := r.db.Model(&User{}).
		Where("id = ?", id).
		Update("traffic_used", 0)

	if result.Error != nil {
		return fmt.Errorf("failed to reset traffic: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
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
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}
	return count, nil
}

// CountExpiredUsers возвращает количество пользователей с истекшим сроком
func (r *Repository) CountExpiredUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).
		Where("expires_at IS NOT NULL").
		Where("expires_at <= ?", time.Now()).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count expired users: %w", err)
	}
	return count, nil
}

// CountUsersOverLimit возвращает количество пользователей, превысивших лимит
func (r *Repository) CountUsersOverLimit() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).
		Where("traffic_limit > 0").
		Where("traffic_used >= traffic_limit").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users over limit: %w", err)
	}
	return count, nil
}
