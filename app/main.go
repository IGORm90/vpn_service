package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vpn-service/api"
	"vpn-service/controllers"
	"vpn-service/database"
	"vpn-service/monitoring"
	"vpn-service/services"
	"vpn-service/xray"

	// Импорты для регистрации компонентов Xray
	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/log"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	_ "github.com/xtls/xray-core/app/router"
	_ "github.com/xtls/xray-core/app/stats"
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/dokodemo"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/proxy/vless/inbound"
	_ "github.com/xtls/xray-core/proxy/vless/outbound"
	_ "github.com/xtls/xray-core/transport/internet/reality"
	_ "github.com/xtls/xray-core/transport/internet/tcp"
)

func main() {
	log.Println("Starting VPN Service with embedded Xray...")

	// Конфигурация из переменных окружения
	dbPath := getEnv("DB_PATH", "./data/vpn.db")
	logPath := getEnv("LOG_PATH", "/var/log/xray/access.log")
	serverPort := getEnv("SERVER_PORT", "8080")
	xrayPrivateKey := getEnv("XRAY_PRIVATE_KEY", "")

	if xrayPrivateKey == "" {
		log.Fatal("XRAY_PRIVATE_KEY environment variable is required")
	}

	// Инициализация базы данных
	log.Println("Initializing database...")
	if err := database.InitDatabase(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	repo := database.NewRepository(database.GetDB())

	// Конфигурация Xray
	xrayConfig := &xray.Config{
		Port:               443,
		RealityPrivateKey:  xrayPrivateKey,
		RealityDest:        getEnv("XRAY_REALITY_DEST", "www.google.com:443"),
		RealityServerNames: []string{getEnv("XRAY_REALITY_SNI", "www.google.com")},
		RealityShortIds:    []string{"", "0123456789abcdef"},
		XHTTPPath:          getEnv("XRAY_XHTTP_PATH", "/xhttp"),
		LogLevel:           getEnv("XRAY_LOG_LEVEL", "info"),
		AccessLogPath:      logPath,
		ErrorLogPath:       getEnv("XRAY_ERROR_LOG", "/var/log/xray/error.log"),
		StatsPort:          10085,
	}

	// Создание менеджера Xray
	log.Println("Initializing Xray manager...")
	xrayManager, err := xray.NewManager(xrayConfig)
	if err != nil {
		log.Fatalf("Failed to create Xray manager: %v", err)
	}

	// Загружаем пользователей и запускаем Xray
	log.Println("Starting Xray server...")
	users, err := repo.ListUsers()
	if err != nil {
		log.Fatalf("Failed to load users: %v", err)
	}

	if err := xrayManager.Start(users); err != nil {
		log.Fatalf("Failed to start Xray: %v", err)
	}
	defer xrayManager.Stop()

	log.Printf("Xray started with %d users", len(users))

	// Инициализация метрик Prometheus
	log.Println("Initializing Prometheus metrics...")
	metrics := monitoring.NewMetrics()
	metricsCollector := monitoring.NewMetricsCollector(metrics, repo)
	metricsCollector.Start(15 * time.Second)
	defer metricsCollector.Stop()

	// Запуск мониторинга логов
	log.Println("Starting log monitor...")
	logMonitor := monitoring.NewLogMonitor(logPath, repo, 30*time.Second)
	if err := logMonitor.Start(); err != nil {
		log.Printf("Warning: failed to start log monitor: %v", err)
	}
	defer logMonitor.Stop()

	// Создание сервисов
	serverIP := getEnv("SERVER_IP", "YOUR_SERVER_IP")
	userService := services.NewUserService(repo, xrayManager, xrayConfig, serverIP)

	// Создание контроллеров
	userController := controllers.NewUserController(userService)

	// Создание API обработчиков и настройка маршрутизатора
	handler := api.NewHandler(repo, xrayManager, xrayConfig)
	router := api.SetupRouter(handler, userController)

	// Запуск HTTP сервера
	server := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("HTTP server listening on port %s", serverPort)
		log.Printf("API documentation:")
		log.Printf("  - POST   /api/users                  - Create user")
		log.Printf("  - GET    /api/users                  - List users")
		log.Printf("  - GET    /api/users/{id}             - Get user")
		log.Printf("  - PATCH  /api/users/{id}             - Update user")
		log.Printf("  - DELETE /api/users/{id}             - Delete user")
		log.Printf("  - GET    /api/users/{id}/config      - Get client config")
		log.Printf("  - POST   /api/users/{id}/reset-traffic - Reset traffic")
		log.Printf("  - GET    /health                     - Health check")
		log.Printf("  - GET    /stats                      - Service stats")
		log.Printf("  - GET    /metrics                    - Prometheus metrics")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	// Останавливаем Xray
	if err := xrayManager.Stop(); err != nil {
		log.Printf("Error stopping Xray: %v", err)
	}

	log.Println("Server stopped")
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
