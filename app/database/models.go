package database

import (
	"time"
)

// User представляет VPN пользователя
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	UUID         string    `gorm:"uniqueIndex;not null" json:"uuid"`
	Secret       string    `json:"secret"` // для будущего Shadowsocks
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	ExpiresAt    time.Time `json:"expires_at"`
	TrafficLimit int64     `gorm:"default:0" json:"traffic_limit"` // 0 = unlimited
	TrafficUsed  int64     `gorm:"default:0" json:"traffic_used"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserDevice представляет устройство пользователя (по IP)
type UserDevice struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;uniqueIndex:idx_user_ip;not null" json:"user_id"`
	IP        string    `gorm:"size:64;uniqueIndex:idx_user_ip;not null" json:"ip"`
	LastSeen  time.Time `gorm:"index" json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsExpired проверяет, истек ли срок действия пользователя
func (u *User) IsExpired() bool {
	return !u.ExpiresAt.IsZero() && time.Now().After(u.ExpiresAt)
}

// IsOverLimit проверяет, превышен ли лимит трафика
func (u *User) IsOverLimit() bool {
	return u.TrafficLimit > 0 && u.TrafficUsed >= u.TrafficLimit
}

// CanConnect проверяет, может ли пользователь подключиться
func (u *User) CanConnect() bool {
	return u.IsActive && !u.IsExpired() && !u.IsOverLimit()
}

// RemainingTraffic возвращает остаток трафика в байтах
func (u *User) RemainingTraffic() int64 {
	if u.TrafficLimit == 0 {
		return -1 // unlimited
	}
	remaining := u.TrafficLimit - u.TrafficUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}
