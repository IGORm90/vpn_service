package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode генерирует QR код для строки и возвращает его в base64
func GenerateQRCode(content string) (string, error) {
	// Генерируем QR код размером 256x256 пикселей
	qrCode, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Конвертируем в base64
	encoded := base64.StdEncoding.EncodeToString(qrCode)
	return encoded, nil
}

// GenerateQRCodePNG генерирует QR код и возвращает его как PNG байты
func GenerateQRCodePNG(content string) ([]byte, error) {
	qrCode, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return qrCode, nil
}
