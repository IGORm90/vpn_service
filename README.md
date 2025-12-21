# VPN Service with Xray, Go, Prometheus & Grafana

A complete VPN service infrastructure with user management, metrics collection, and monitoring.

## 📁 Project Structure

```
project/
├── docker-compose.yml      # Orchestrates all services
├── go-app/
│   ├── Dockerfile         # Go service container
│   ├── main.go           # User management & metrics service
│   └── go.mod            # Go dependencies
├── xray/
│   └── config.json       # VLESS+xHTTP configuration
├── prometheus/
│   └── prometheus.yml    # Metrics collection config
└── grafana/
    ├── datasources/
    │   └── datasource.yml
    └── dashboards/
        ├── dashboard.yml
        └── vpn-metrics.json
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Xray client for testing

### 1. Generate Xray Keys

Before starting, generate Reality keys for Xray:

```bash
docker run --rm teddysun/xray:latest xray x25519
```

Replace `GENERATE_YOUR_PRIVATE_KEY_HERE` in `xray/config.json` with your generated private key.

### 2. Start Services

```bash
docker-compose up -d
```

### 3. Access Services

- **Go API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Xray VPN**: Port 443 (VLESS+Reality)

## 📊 Monitoring

### Grafana Dashboard

1. Open Grafana at http://localhost:3000
2. Login with `admin`/`admin`
3. Navigate to Dashboards → VPN Service Metrics
4. View real-time VPN metrics:
   - Active users
   - Connection rates
   - Bandwidth usage
   - Total connections

### Prometheus Metrics

View raw metrics at: http://localhost:8080/metrics

Available metrics:
- `vpn_active_users` - Number of active users
- `vpn_total_connections` - Total connections count
- `vpn_bandwidth_bytes{user_id, direction}` - Bandwidth per user
- `vpn_connection_duration_seconds` - Connection duration histogram

## 🔧 API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Create User
```bash
curl -X POST http://localhost:8080/api/users/create \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "uuid": "b831381d-6324-4d53-ad4f-8cda48b30811"
  }'
```

### List Users
```bash
curl http://localhost:8080/api/users/list
```

### Get User
```bash
curl http://localhost:8080/api/users/get?id=user_1234567890
```

### Deactivate User
```bash
curl -X POST http://localhost:8080/api/users/deactivate?id=user_1234567890
```

## 🔐 Xray Configuration

The service uses **VLESS + Reality + xHTTP** protocol for maximum security and censorship resistance.

### Client Configuration

```json
{
  "protocol": "vless",
  "settings": {
    "vnext": [{
      "address": "YOUR_SERVER_IP",
      "port": 443,
      "users": [{
        "id": "b831381d-6324-4d53-ad4f-8cda48b30811",
        "encryption": "none",
        "flow": ""
      }]
    }]
  },
  "streamSettings": {
    "network": "xhttp",
    "security": "reality",
    "realitySettings": {
      "serverName": "www.google.com",
      "fingerprint": "chrome",
      "publicKey": "YOUR_PUBLIC_KEY",
      "shortId": "0123456789abcdef",
      "spiderX": ""
    },
    "xhttpSettings": {
      "path": "/xhttp",
      "host": "www.google.com"
    }
  }
}
```

## 📝 Configuration Files

### Customize Xray Users

Edit `xray/config.json` and add users in the `clients` array:

```json
"clients": [
  {
    "id": "UUID_HERE",
    "email": "user@example.com",
    "flow": ""
  }
]
```

Generate UUIDs: `uuidgen` (macOS/Linux) or online UUID generator

### Adjust Prometheus Scrape Intervals

Edit `prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s  # Adjust as needed
```

### Grafana Dashboard Customization

Import additional dashboards:
1. Go to Grafana
2. Click + → Import
3. Enter dashboard ID or paste JSON

## 🛠️ Development

### Rebuild Go Service

```bash
docker-compose build go-app
docker-compose up -d go-app
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f go-app
docker-compose logs -f xray
```

### Stop Services

```bash
docker-compose down
```

### Clean All Data

```bash
docker-compose down -v
```

## 🔍 Troubleshooting

### Xray Not Starting

1. Check if private key is set in `xray/config.json`
2. Verify port 443 is not in use: `lsof -i :443`
3. Check logs: `docker-compose logs xray`

### Go Service Can't Connect to Prometheus

1. Ensure all services are in the same Docker network
2. Check `docker-compose ps` - all services should be "Up"
3. Restart: `docker-compose restart`

### Grafana Shows No Data

1. Check Prometheus is collecting metrics: http://localhost:9090/targets
2. Verify Go app metrics endpoint: http://localhost:8080/metrics
3. Check Grafana datasource connection: Configuration → Data Sources

## 📈 Scaling

To scale the Go service:

```bash
docker-compose up -d --scale go-app=3
```

Add a load balancer (nginx/traefik) in front of multiple instances.

## 🔒 Security Notes

1. **Change default passwords** in `docker-compose.yml` (Grafana)
2. **Generate unique keys** for Xray Reality
3. **Use HTTPS** for Go API in production (add reverse proxy)
4. **Implement authentication** for API endpoints
5. **Restrict ports** using firewall rules
6. **Regular updates**: `docker-compose pull && docker-compose up -d`

## 📚 Technologies

- **Xray-core**: Modern VPN proxy with Reality protocol
- **Go**: High-performance user management service
- **Prometheus**: Metrics collection and storage
- **Grafana**: Beautiful metrics visualization
- **Docker**: Containerized deployment

## 📄 License

MIT License - Use freely for personal and commercial projects.

## 🤝 Contributing

Contributions welcome! Please submit pull requests or open issues.

---

**Happy VPN-ing! 🚀**

