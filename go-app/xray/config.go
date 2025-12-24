package xray

import (
	"encoding/json"
	"fmt"
	"os"
	"vpn-service/database"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

// Config содержит параметры конфигурации Xray
type Config struct {
	Port               int
	RealityPrivateKey  string
	RealityDest        string
	RealityServerNames []string
	RealityShortIds    []string
	XHTTPPath          string
	LogLevel           string
	AccessLogPath      string
	ErrorLogPath       string
	StatsPort          int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Port:               443,
		RealityPrivateKey:  os.Getenv("XRAY_PRIVATE_KEY"),
		RealityDest:        "www.google.com:443",
		RealityServerNames: []string{"www.google.com"},
		RealityShortIds:    []string{"", "0123456789abcdef"},
		XHTTPPath:          "/xhttp",
		LogLevel:           "info",
		AccessLogPath:      "/var/log/xray/access.log",
		ErrorLogPath:       "/var/log/xray/error.log",
		StatsPort:          10085,
	}
}

// GenerateConfig генерирует конфигурацию Xray из списка пользователей
func GenerateConfig(users []*database.User, cfg *Config) (*core.Config, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Создаем JSON конфигурацию
	configJSON := map[string]interface{}{
		"log": map[string]interface{}{
			"access":   cfg.AccessLogPath,
			"error":    cfg.ErrorLogPath,
			"loglevel": cfg.LogLevel,
		},
		"api": map[string]interface{}{
			"tag": "api",
			"services": []string{
				"HandlerService",
				"LoggerService",
				"StatsService",
			},
		},
		"stats": map[string]interface{}{},
		"policy": map[string]interface{}{
			"levels": map[string]interface{}{
				"0": map[string]interface{}{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]interface{}{
				"statsInboundUplink":    true,
				"statsInboundDownlink":  true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		},
		"inbounds": generateInbounds(users, cfg),
		"outbounds": []map[string]interface{}{
			{
				"protocol": "freedom",
				"tag":      "direct",
				"settings": map[string]interface{}{},
			},
			{
				"protocol": "blackhole",
				"tag":      "block",
				"settings": map[string]interface{}{},
			},
		},
		"routing": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"type":        "field",
					"inboundTag":  []string{"api-in"},
					"outboundTag": "api",
				},
				{
					"type":        "field",
					"protocol":    []string{"bittorrent"},
					"outboundTag": "block",
				},
			},
		},
	}

	// Конвертируем в JSON и обратно через conf парсер
	jsonBytes, err := json.Marshal(configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Парсим через xray-core конфигуратор
	config := &conf.Config{}
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Конвертируем в protobuf конфигурацию
	pbConfig, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	return pbConfig, nil
}

// generateInbounds генерирует список inbounds для конфигурации
func generateInbounds(users []*database.User, cfg *Config) []map[string]interface{} {
	// Генерируем список клиентов из активных пользователей
	clients := make([]map[string]interface{}, 0)
	for _, user := range users {
		if user.CanConnect() {
			clients = append(clients, map[string]interface{}{
				"id":    user.UUID,
				"email": user.Username,
				"flow":  "",
			})
		}
	}

	return []map[string]interface{}{
		{
			"port":     cfg.Port,
			"protocol": "vless",
			"tag":      "vless-in",
			"settings": map[string]interface{}{
				"clients":    clients,
				"decryption": "none",
			},
			"streamSettings": map[string]interface{}{
				"network":  "xhttp",
				"security": "reality",
				"realitySettings": map[string]interface{}{
					"show":        false,
					"dest":        cfg.RealityDest,
					"xver":        0,
					"serverNames": cfg.RealityServerNames,
					"privateKey":  cfg.RealityPrivateKey,
					"shortIds":    cfg.RealityShortIds,
				},
				"xhttpSettings": map[string]interface{}{
					"mode": "auto",
					"path": cfg.XHTTPPath,
					"host": cfg.RealityServerNames[0],
				},
			},
			"sniffing": map[string]interface{}{
				"enabled": true,
				"destOverride": []string{
					"http",
					"tls",
				},
			},
		},
		{
			"listen":   "0.0.0.0",
			"port":     cfg.StatsPort,
			"protocol": "dokodemo-door",
			"tag":      "api-in",
			"settings": map[string]interface{}{
				"address": "127.0.0.1",
			},
		},
	}
}

// ValidateConfig проверяет корректность конфигурации
func ValidateConfig(cfg *Config) error {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}

	if cfg.RealityPrivateKey == "" {
		return fmt.Errorf("reality private key is required")
	}

	if cfg.RealityDest == "" {
		return fmt.Errorf("reality destination is required")
	}

	if len(cfg.RealityServerNames) == 0 {
		return fmt.Errorf("at least one reality server name is required")
	}

	return nil
}
