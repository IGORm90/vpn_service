# VPN Service with VLESS Protocol

A complete VPN service infrastructure with embedded Xray-core, user management, traffic monitoring, and metrics collection.

> **🇷🇺 Русская версия:** [README.ru.md](README.ru.md)

## 🏗️ Architecture

- **Go Application** - Embedded Xray-core, REST API, User Management
- **SQLite** - User database
- **Prometheus** - Metrics collection
- **Grafana** - Metrics visualization
- **VLESS + Reality + xHTTP** - Modern secure VPN protocol

## ✨ Features

- ✅ REST API for user management
- ✅ Traffic limits and subscription expiry
- ✅ Automatic traffic tracking
- ✅ Client config generation (JSON, URI, QR code)
- ✅ Real-time monitoring (Prometheus + Grafana)
- ✅ VLESS protocol with Reality and xHTTP
- ✅ SQLite database for user storage

## 📁 Project Structure

```
vpn-service/
├── go-app/
│   ├── main.go              # Entry point
│   ├── database/            # Database layer
│   │   ├── models.go        # User model
│   │   ├── database.go      # DB connection
│   │   └── repository.go    # CRUD operations
│   ├── xray/                # Xray integration
│   │   ├── manager.go       # Xray instance management
│   │   ├── config.go        # Config generation
│   │   └── client_config.go # Client configs
│   ├── api/                 # REST API
│   │   ├── handlers.go      # HTTP handlers
│   │   ├── middleware.go    # Middleware
│   │   └── responses.go     # JSON responses
│   ├── monitoring/          # Monitoring
│   │   ├── log_parser.go    # Log parser
│   │   └── metrics.go       # Prometheus metrics
│   └── utils/               # Utilities
│       ├── crypto.go        # Password hashing
│       └── qrcode.go        # QR code generation
├── prometheus/
│   └── prometheus.yml       # Prometheus config
├── grafana/
│   ├── datasources/         # Data sources
│   └── dashboards/          # Dashboards
├── examples/                # Usage examples
│   ├── api_examples.sh      # API examples
│   ├── create_test_users.sh # Test users
│   └── python_client.py     # Python client
├── docker-compose.yml       # Docker Compose
├── Makefile                 # Automation
└── README.md               # Documentation
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- make (optional, for convenience)

### 1. Setup

```bash
# Clone the repository
git clone <repo-url>
cd vpn-service

# Generate keys and create .env
make setup

# Edit .env file
nano .env
```

Set required variables:
- `XRAY_PRIVATE_KEY` - Private key from `make generate-keys`
- `SERVER_IP` - Your server IP address

### 2. Start Services

```bash
# Build and start
make build
make up

# Check status
make status

# View logs
make logs
```

### 3. Create First User

```bash
# Using Makefile
make create-user

# Or directly with curl
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "secret123",
    "traffic_limit": 10737418240,
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

### 4. Access Services

- **API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **VLESS VPN**: Port 443

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

### Users

#### Create User
```bash
POST /api/users
Content-Type: application/json

{
  "username": "john_doe",
  "password": "securepass123",
  "traffic_limit": 10737418240,  // 10GB in bytes, 0 = unlimited
  "expires_at": "2025-12-31T23:59:59Z"  // optional
}
```

#### List Users
```bash
GET /api/users
GET /api/users?active=true  // only active users
```

#### Get User
```bash
GET /api/users/{id}
```

#### Update User
```bash
PATCH /api/users/{id}
Content-Type: application/json

{
  "traffic_limit": 21474836480,  // 20GB
  "is_active": true
}
```

#### Delete User
```bash
DELETE /api/users/{id}
```

#### Get User Config
```bash
GET /api/users/{id}/config
```

Returns:
- JSON config for v2rayN, Nekoray
- VLESS URI
- QR code (base64)
- Traffic statistics

#### Reset Traffic
```bash
POST /api/users/{id}/reset-traffic
```

### System

#### Health Check
```bash
GET /health
```

#### Statistics
```bash
GET /stats
```

#### Prometheus Metrics
```bash
GET /metrics
```

## 🔐 Protocol Configuration

The service uses **VLESS + Reality + xHTTP** protocol for maximum security and censorship resistance.

Client configuration is generated automatically via API:
```bash
GET /api/users/{id}/config
```

This returns JSON config, VLESS URI, and QR code ready for import into VPN clients.

## 🛠️ Makefile Commands

### Main Commands
```bash
make help          # Show help
make setup         # Initial setup
make up            # Start services
make down          # Stop services
make restart       # Restart services
make logs          # View logs
make status        # Service status
make clean         # Remove all data
```

### User Management
```bash
make create-user                    # Create test user
make list-users                     # List users
make get-user-config USER_ID=1      # Get config
make delete-user USER_ID=1          # Delete user
```

### Testing
```bash
make test-health   # Check health
make test-stats    # Check statistics
make metrics       # Show metrics
```

### Backup
```bash
make db-backup     # Backup database
make db-restore    # Restore database
```

## 📱 Client Applications

### Android
- v2rayNG
- Nekoray

### iOS
- Shadowrocket
- V2Box

### Windows/macOS/Linux
- v2rayN / v2rayNG
- Nekoray
- Qv2ray

### Connection
1. Get config: `GET /api/users/{id}/config`
2. Use QR code or VLESS URI
3. Import into client application

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

