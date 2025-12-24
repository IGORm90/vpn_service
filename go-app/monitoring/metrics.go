package monitoring

import (
	"log"
	"time"
	"vpn-service/database"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics содержит все метрики Prometheus
type Metrics struct {
	ActiveUsers      prometheus.Gauge
	TotalUsers       prometheus.Gauge
	ExpiredUsers     prometheus.Gauge
	UsersOverLimit   prometheus.Gauge
	TotalTraffic     *prometheus.CounterVec
	UserTraffic      *prometheus.GaugeVec
	ConnectionsTotal prometheus.Counter
	ConnectionActive prometheus.Gauge
	UserLimitRemain  *prometheus.GaugeVec
}

// NewMetrics создает и регистрирует метрики
func NewMetrics() *Metrics {
	m := &Metrics{
		ActiveUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_active_users_total",
			Help: "Number of active VPN users (not expired, not over limit)",
		}),
		TotalUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_total_users",
			Help: "Total number of VPN users",
		}),
		ExpiredUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_expired_users",
			Help: "Number of users with expired subscriptions",
		}),
		UsersOverLimit: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_users_over_limit",
			Help: "Number of users over traffic limit",
		}),
		TotalTraffic: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vpn_traffic_bytes_total",
				Help: "Total traffic in bytes",
			},
			[]string{"direction"},
		),
		UserTraffic: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_user_traffic_bytes",
				Help: "Traffic usage per user in bytes",
			},
			[]string{"username", "uuid", "direction"},
		),
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_connections_total",
			Help: "Total number of VPN connections established",
		}),
		ConnectionActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_connections_active",
			Help: "Number of active VPN connections",
		}),
		UserLimitRemain: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_user_limit_remaining_bytes",
				Help: "Remaining traffic limit for user in bytes",
			},
			[]string{"username", "uuid"},
		),
	}

	// Регистрируем все метрики
	prometheus.MustRegister(m.ActiveUsers)
	prometheus.MustRegister(m.TotalUsers)
	prometheus.MustRegister(m.ExpiredUsers)
	prometheus.MustRegister(m.UsersOverLimit)
	prometheus.MustRegister(m.TotalTraffic)
	prometheus.MustRegister(m.UserTraffic)
	prometheus.MustRegister(m.ConnectionsTotal)
	prometheus.MustRegister(m.ConnectionActive)
	prometheus.MustRegister(m.UserLimitRemain)

	return m
}

// MetricsCollector собирает метрики из базы данных
type MetricsCollector struct {
	metrics    *Metrics
	repository *database.Repository
	stopCh     chan struct{}
	running    bool
}

// NewMetricsCollector создает новый коллектор метрик
func NewMetricsCollector(metrics *Metrics, repo *database.Repository) *MetricsCollector {
	return &MetricsCollector{
		metrics:    metrics,
		repository: repo,
		stopCh:     make(chan struct{}),
		running:    false,
	}
}

// Start запускает периодический сбор метрик
func (c *MetricsCollector) Start(interval time.Duration) {
	if c.running {
		return
	}

	c.running = true

	go func() {
		// Первый сбор сразу
		c.collectMetrics()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.collectMetrics()
			case <-c.stopCh:
				return
			}
		}
	}()

	log.Printf("Metrics collector started (interval: %v)", interval)
}

// Stop останавливает сбор метрик
func (c *MetricsCollector) Stop() {
	if !c.running {
		return
	}

	close(c.stopCh)
	c.running = false
	log.Println("Metrics collector stopped")
}

// collectMetrics собирает все метрики из БД
func (c *MetricsCollector) collectMetrics() {
	// Подсчитываем пользователей
	totalUsers, err := c.repository.CountUsers()
	if err != nil {
		log.Printf("Failed to count total users: %v", err)
	} else {
		c.metrics.TotalUsers.Set(float64(totalUsers))
	}

	activeUsers, err := c.repository.CountActiveUsers()
	if err != nil {
		log.Printf("Failed to count active users: %v", err)
	} else {
		c.metrics.ActiveUsers.Set(float64(activeUsers))
	}

	expiredUsers, err := c.repository.CountExpiredUsers()
	if err != nil {
		log.Printf("Failed to count expired users: %v", err)
	} else {
		c.metrics.ExpiredUsers.Set(float64(expiredUsers))
	}

	overLimitUsers, err := c.repository.CountUsersOverLimit()
	if err != nil {
		log.Printf("Failed to count users over limit: %v", err)
	} else {
		c.metrics.UsersOverLimit.Set(float64(overLimitUsers))
	}

	// Собираем метрики по каждому пользователю
	users, err := c.repository.ListUsers()
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return
	}

	var totalUpload, totalDownload int64
	for _, user := range users {
		// Предполагаем что трафик примерно 50/50 upload/download
		// В реальности нужно хранить отдельно
		upload := user.TrafficUsed / 2
		download := user.TrafficUsed / 2

		c.metrics.UserTraffic.WithLabelValues(
			user.Username, user.UUID, "upload",
		).Set(float64(upload))

		c.metrics.UserTraffic.WithLabelValues(
			user.Username, user.UUID, "download",
		).Set(float64(download))

		// Остаток лимита
		remaining := user.RemainingTraffic()
		if remaining >= 0 {
			c.metrics.UserLimitRemain.WithLabelValues(
				user.Username, user.UUID,
			).Set(float64(remaining))
		}

		totalUpload += upload
		totalDownload += download
	}

	// Обновляем общий трафик (используем Add только для новых данных)
	// Здесь используем Set через Gauge если нужно точное значение
}

// UpdateConnection обновляет метрики подключений
func (c *MetricsCollector) UpdateConnection(active int) {
	c.metrics.ConnectionActive.Set(float64(active))
}

// IncrementConnections увеличивает счетчик подключений
func (c *MetricsCollector) IncrementConnections() {
	c.metrics.ConnectionsTotal.Inc()
}
