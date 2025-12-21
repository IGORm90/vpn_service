package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// User represents a VPN user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// UserService manages VPN users
type UserService struct {
	users map[string]*User
	mu    sync.RWMutex
}

// Metrics collectors
var (
	activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vpn_active_users",
		Help: "Number of active VPN users",
	})

	totalConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "vpn_total_connections",
		Help: "Total number of VPN connections",
	})

	bandwidthUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vpn_bandwidth_bytes",
			Help: "Bandwidth usage in bytes",
		},
		[]string{"user_id", "direction"},
	)

	connectionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "vpn_connection_duration_seconds",
		Help:    "Duration of VPN connections",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	// Register metrics
	prometheus.MustRegister(activeUsers)
	prometheus.MustRegister(totalConnections)
	prometheus.MustRegister(bandwidthUsage)
	prometheus.MustRegister(connectionDuration)
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

// CreateUser creates a new VPN user
func (s *UserService) CreateUser(email, uuid string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &User{
		ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
		Email:     email,
		UUID:      uuid,
		CreatedAt: time.Now(),
		Active:    true,
	}

	s.users[user.ID] = user
	activeUsers.Inc()

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// ListUsers returns all users
func (s *UserService) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

// DeactivateUser deactivates a user
func (s *UserService) DeactivateUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if user.Active {
		user.Active = false
		activeUsers.Dec()
	}

	return nil
}

// Handler functions
func (s *UserService) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
		UUID  string `json:"uuid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.CreateUser(req.Email, req.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalConnections.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	user, err := s.GetUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users := s.ListUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *UserService) handleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	if err := s.DeactivateUser(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s deactivated", id)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	port := os.Getenv("PROMETHEUS_PORT")
	if port == "" {
		port = "8080"
	}

	userService := NewUserService()

	// API routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/api/users/create", userService.handleCreateUser)
	http.HandleFunc("/api/users/get", userService.handleGetUser)
	http.HandleFunc("/api/users/list", userService.handleListUsers)
	http.HandleFunc("/api/users/deactivate", userService.handleDeactivateUser)

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start metrics simulation (for demo purposes)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Simulate bandwidth usage
			bandwidthUsage.WithLabelValues("user_demo", "upload").Add(1024 * 1024)
			bandwidthUsage.WithLabelValues("user_demo", "download").Add(5 * 1024 * 1024)
		}
	}()

	log.Printf("VPN Service starting on port %s", port)
	log.Printf("Metrics available at http://localhost:%s/metrics", port)
	log.Printf("API endpoints:")
	log.Printf("  - POST /api/users/create")
	log.Printf("  - GET  /api/users/get?id=<user_id>")
	log.Printf("  - GET  /api/users/list")
	log.Printf("  - POST /api/users/deactivate?id=<user_id>")
	log.Printf("  - GET  /health")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
