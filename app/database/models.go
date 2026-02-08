package database

import (
	"time"
)

// User представляет VPN пользователя
type User struct {
	ID        int64     `gorm:"uniqueIndex;not null" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	UUID      string    `gorm:"uniqueIndex;not null" json:"uuid"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDevice представляет устройство пользователя (по IP)
type UserDevice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"index;uniqueIndex:idx_user_ip;not null" json:"user_id"`
	IP        string    `gorm:"size:64;uniqueIndex:idx_user_ip;not null" json:"ip"`
	LastSeen  time.Time `gorm:"index" json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanConnect проверяет, может ли пользователь подключиться
func (u *User) CanConnect() bool {
	return u.IsActive
}
