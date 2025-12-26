package api

import (
	"vpn-service/controllers"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter настраивает и возвращает настроенный маршрутизатор
func SetupRouter(handler *Handler, userController *controllers.UserController) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(LoggingMiddleware)
	router.Use(RecoveryMiddleware)
	router.Use(CORSMiddleware)

	// API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Users - используем контроллер
	apiRouter.HandleFunc("/users", userController.CreateUser).Methods("POST")
	apiRouter.HandleFunc("/users", userController.ListUsers).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userController.GetUser).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userController.UpdateUser).Methods("PATCH", "PUT")
	apiRouter.HandleFunc("/users/{id}", userController.DeleteUser).Methods("DELETE")
	apiRouter.HandleFunc("/users/{id}/config", userController.GetUserConfig).Methods("GET")
	apiRouter.HandleFunc("/users/{id}/reset-traffic", userController.ResetTraffic).Methods("POST")

	// System - используем handler для системных endpoints
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/stats", handler.GetStats).Methods("GET")

	// Prometheus metrics
	router.Handle("/metrics", promhttp.Handler())

	// Статические файлы (опционально)
	// router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	return router
}
