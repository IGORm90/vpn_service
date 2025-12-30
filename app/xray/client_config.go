package xray

import (
	"encoding/json"
	"fmt"
	"net/url"
	"vpn-service/database"
)

// ClientConfig представляет конфигурацию клиента
type ClientConfig struct {
	ServerAddress string
	ServerPort    int
	UUID          string
	PublicKey     string
	ShortID       string
	ServerName    string
	XHTTPPath     string
}

// GenerateClientJSON генерирует JSON конфигурацию для клиента
func GenerateClientJSON(user *database.User, cfg *Config, serverIP string) (string, error) {
	if !user.CanConnect() {
		return "", fmt.Errorf("user cannot connect (inactive, expired or over limit)")
	}

	clientConfig := map[string]interface{}{
		"protocol": "vless",
		"settings": map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": serverIP,
					"port":    cfg.Port,
					"users": []map[string]interface{}{
						{
							"id":         user.UUID,
							"encryption": "none",
							"flow":       "",
						},
					},
				},
			},
		},
		"streamSettings": map[string]interface{}{
			"network":  "tcp",
			"security": "reality",
			"realitySettings": map[string]interface{}{
				"serverName":  cfg.RealityServerNames[0],
				"fingerprint": "chrome",
				"publicKey":   cfg.RealityPublicKey,
				"shortId":     cfg.RealityShortIds[1],
				"spiderX":     "",
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(clientConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal client config: %w", err)
	}

	return string(jsonBytes), nil
}

// GenerateVlessURI генерирует VLESS URI для клиента
func GenerateVlessURI(user *database.User, cfg *Config, serverIP string) (string, error) {
	if !user.CanConnect() {
		return "", fmt.Errorf("user cannot connect (inactive, expired or over limit)")
	}

	// Формат: vless://UUID@SERVER:PORT?params#REMARK
	params := url.Values{}
	params.Set("type", "tcp")
	params.Set("security", "reality")
	params.Set("pbk", cfg.RealityPublicKey)
	params.Set("fp", "chrome")
	params.Set("sni", cfg.RealityServerNames[0])
	params.Set("sid", cfg.RealityShortIds[1])

	uri := fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.UUID,
		serverIP,
		cfg.Port,
		params.Encode(),
		url.QueryEscape(user.Username),
	)

	return uri, nil
}

// GenerateShadowrocketURI генерирует URI для Shadowrocket (iOS)
func GenerateShadowrocketURI(user *database.User, cfg *Config, serverIP string) (string, error) {
	// Shadowrocket использует тот же формат что и обычный VLESS URI
	return GenerateVlessURI(user, cfg, serverIP)
}

// GenerateClashConfig генерирует конфигурацию для Clash
func GenerateClashConfig(user *database.User, cfg *Config, serverIP string) (string, error) {
	if !user.CanConnect() {
		return "", fmt.Errorf("user cannot connect (inactive, expired or over limit)")
	}

	clashConfig := map[string]interface{}{
		"proxies": []map[string]interface{}{
			{
				"name":    user.Username,
				"type":    "vless",
				"server":  serverIP,
				"port":    cfg.Port,
				"uuid":    user.UUID,
				"network": "tcp",
				"tls":     true,
				"udp":     true,
				"flow":    "",
				"reality-opts": map[string]interface{}{
					"public-key": cfg.RealityPublicKey,
					"short-id":   cfg.RealityShortIds[1],
				},
			},
		},
	}

	yamlBytes, err := json.MarshalIndent(clashConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal clash config: %w", err)
	}

	return string(yamlBytes), nil
}

// ClientConfigResponse представляет ответ с конфигурациями клиента
type ClientConfigResponse struct {
	Username     string `json:"username"`
	UUID         string `json:"uuid"`
	ServerIP     string `json:"server_ip"`
	ServerPort   int    `json:"server_port"`
	JSON         string `json:"config_json"`
	URI          string `json:"vless_uri"`
	QRCode       string `json:"qr_code,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	TrafficLimit int64  `json:"traffic_limit"`
	TrafficUsed  int64  `json:"traffic_used"`
	IsActive     bool   `json:"is_active"`
}
