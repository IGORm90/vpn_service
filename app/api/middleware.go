package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"vpn-service/responses"
)

// LoggingMiddleware логирует HTTP запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для перехвата статус кода
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("[%s] %s %s - %d (%v)",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter оборачивает http.ResponseWriter для перехвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware добавляет CORS заголовки
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware восстанавливает приложение после паники
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				responses.SendInternalError(w, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware устанавливает Content-Type для JSON
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware проверяет Bearer токен в заголовке Authorization
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из переменной окружения
		expectedToken := os.Getenv("API_BEARER_TOKEN")
		if expectedToken == "" {
			log.Println("Warning: API_BEARER_TOKEN is not set, authentication disabled")
			next.ServeHTTP(w, r)
			return
		}

		// Получаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			responses.SendUnauthorized(w, "Missing authorization header")
			return
		}

		// Проверяем формат Bearer токена
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			responses.SendUnauthorized(w, "Invalid authorization header format. Expected: Bearer <token>")
			return
		}

		// Проверяем токен
		token := parts[1]
		if token != expectedToken {
			responses.SendUnauthorized(w, "Invalid authentication token")
			return
		}

		// Токен валиден, продолжаем обработку запроса
		next.ServeHTTP(w, r)
	})
}
