package api

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRouter настраивает и возвращает настроенный маршрутизатор
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(LoggingMiddleware)
	router.Use(RecoveryMiddleware)
	router.Use(CORSMiddleware)

	// API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Users
	apiRouter.HandleFunc("/users", handler.CreateUser).Methods("POST")
	apiRouter.HandleFunc("/users", handler.ListUsers).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", handler.UpdateUser).Methods("PATCH", "PUT")
	apiRouter.HandleFunc("/users/{id}", handler.DeleteUser).Methods("DELETE")
	apiRouter.HandleFunc("/users/{id}/config", handler.GetUserConfig).Methods("GET")
	apiRouter.HandleFunc("/users/{id}/reset-traffic", handler.ResetTraffic).Methods("POST")

	// System
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/stats", handler.GetStats).Methods("GET")

	// Prometheus metrics
	router.Handle("/metrics", promhttp.Handler())

	// Статические файлы (опционально)
	// router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	return router
}
