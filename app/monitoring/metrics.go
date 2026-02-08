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
	TotalTraffic     *prometheus.CounterVec
	ConnectionsTotal prometheus.Counter
	ConnectionActive prometheus.Gauge
}

// NewMetrics создает и регистрирует метрики
func NewMetrics() *Metrics {
	m := &Metrics{
		ActiveUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_active_users_total",
			Help: "Number of active VPN users",
		}),
		TotalUsers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_total_users",
			Help: "Total number of VPN users",
		}),
		TotalTraffic: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vpn_traffic_bytes_total",
				Help: "Total traffic in bytes",
			},
			[]string{"direction"},
		),
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_connections_total",
			Help: "Total number of VPN connections established",
		}),
		ConnectionActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_connections_active",
			Help: "Number of active VPN connections",
		}),
	}

	// Регистрируем все метрики
	prometheus.MustRegister(m.ActiveUsers)
	prometheus.MustRegister(m.TotalUsers)
	prometheus.MustRegister(m.TotalTraffic)
	prometheus.MustRegister(m.ConnectionsTotal)
	prometheus.MustRegister(m.ConnectionActive)

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

}

// Stop останавливает сбор метрик
func (c *MetricsCollector) Stop() {
	if !c.running {
		return
	}

	close(c.stopCh)
	c.running = false
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
}

// UpdateConnection обновляет метрики подключений
func (c *MetricsCollector) UpdateConnection(active int) {
	c.metrics.ConnectionActive.Set(float64(active))
}

// IncrementConnections увеличивает счетчик подключений
func (c *MetricsCollector) IncrementConnections() {
	c.metrics.ConnectionsTotal.Inc()
}
