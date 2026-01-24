# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Установка зависимостей для сборки
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o vpn-service .

# Final stage
FROM alpine:latest

# Установка runtime зависимостей
RUN apk --no-cache add ca-certificates tzdata sqlite-libs iptables

# Создание директорий
RUN mkdir -p /app/data /var/log/xray

# Включение IP forwarding (будет применено при запуске с NET_ADMIN)
RUN echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/vpn-service .

# Expose ports
# 443 - VLESS
# 8080 - API
# 10085 - Xray Stats API
EXPOSE 443 8080 10085

# Создание volume точек
VOLUME ["/app/data", "/var/log/xray"]

CMD ["./vpn-service"]

