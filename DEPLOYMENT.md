# Deployment Guide

Руководство по развертыванию VPN сервиса на production сервере.

## Требования к серверу

### Минимальные
- **OS**: Ubuntu 20.04+ / Debian 11+ / CentOS 8+
- **RAM**: 1GB
- **CPU**: 1 core
- **Disk**: 10GB
- **Network**: Публичный IP адрес

### Рекомендуемые
- **RAM**: 2GB+
- **CPU**: 2+ cores
- **Disk**: 20GB+ SSD
- **Network**: 100 Mbps+

## Подготовка сервера

### 1. Обновление системы

```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade -y

# CentOS/RHEL
sudo yum update -y
```

### 2. Установка Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Перелогиньтесь для применения изменений
```

### 3. Установка Docker Compose

```bash
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 4. Настройка Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 443/tcp   # VLESS
sudo ufw allow 8080/tcp  # API (опционально, для удаленного доступа)
sudo ufw enable

# firewalld (CentOS)
sudo firewall-cmd --permanent --add-port=22/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

## Развертывание

### 1. Клонирование репозитория

```bash
cd /opt
sudo git clone <repository-url> vpn-service
cd vpn-service
sudo chown -R $USER:$USER .
```

### 2. Конфигурация

```bash
# Генерация ключей
make generate-keys

# Создание .env
cp .env.example .env
nano .env
```

Установите в `.env`:
```bash
XRAY_PRIVATE_KEY=<your_generated_private_key>
SERVER_IP=<your_server_public_ip>
```

### 3. Сборка и запуск

```bash
# Сборка образов
make build

# Запуск сервисов
make up

# Проверка статуса
make status
```

### 4. Проверка работоспособности

```bash
# Health check
curl http://localhost:8080/health

# Создание тестового пользователя
make create-user

# Просмотр логов
make logs
```

## Настройка SSL/TLS для API (Опционально)

Для защиты API рекомендуется использовать reverse proxy (nginx/traefik) с SSL сертификатом.

### Использование Nginx

```bash
# Установка nginx
sudo apt install nginx certbot python3-certbot-nginx -y

# Конфигурация
sudo nano /etc/nginx/sites-available/vpn-api
```

Конфигурация nginx:
```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Активация и SSL:
```bash
sudo ln -s /etc/nginx/sites-available/vpn-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Получение SSL сертификата
sudo certbot --nginx -d api.yourdomain.com
```

## Мониторинг

### Grafana

1. Откройте http://your-server-ip:3000
2. Логин: admin / admin (измените пароль!)
3. Импортируйте дашборды из `grafana/dashboards/`

### Prometheus

Доступ: http://your-server-ip:9090

### Логи

```bash
# Все логи
make logs

# Только Go app
make logs-go

# Логи Xray
docker exec vpn-go-app tail -f /var/log/xray/access.log
```

## Резервное копирование

### Автоматический backup базы данных

Создайте cron job:
```bash
crontab -e
```

Добавьте:
```bash
# Backup каждый день в 3:00
0 3 * * * cd /opt/vpn-service && make db-backup && cp vpn.db.backup /backups/vpn-$(date +\%Y\%m\%d).db
```

### Ручной backup

```bash
# Backup БД
make db-backup

# Копирование backup
cp vpn.db.backup /path/to/safe/location/
```

### Восстановление

```bash
# Копируем backup
cp /path/to/backup/vpn.db.backup .

# Восстанавливаем
make db-restore
```

## Обновление

```bash
# Остановка сервисов
make down

# Backup БД
make db-backup

# Обновление кода
git pull

# Пересборка
make rebuild

# Запуск
make up

# Проверка
make test-health
```

## Безопасность

### 1. Изменение паролей по умолчанию

```yaml
# docker-compose.yml
grafana:
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=your_strong_password
```

### 2. Ограничение доступа к API

Используйте firewall для ограничения доступа к API:
```bash
# Разрешить API только с определенных IP
sudo ufw allow from YOUR_ADMIN_IP to any port 8080
```

### 3. Регулярные обновления

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Обновление Docker образов
docker-compose pull
make rebuild
```

### 4. Мониторинг логов

Настройте мониторинг на подозрительную активность:
```bash
# Просмотр попыток подключений
docker exec vpn-go-app grep "connection" /var/log/xray/access.log
```

## Масштабирование

### Вертикальное масштабирование

Увеличьте ресурсы сервера (CPU, RAM).

### Горизонтальное масштабирование

1. Настройте несколько серверов
2. Используйте общую базу данных (PostgreSQL)
3. Настройте load balancer (nginx, HAProxy)
4. Синхронизируйте конфигурацию между серверами

## Troubleshooting

### Контейнер не запускается

```bash
# Проверка логов
docker logs vpn-go-app

# Проверка конфигурации
docker-compose config
```

### Порт 443 занят

```bash
# Проверка что использует порт
sudo lsof -i :443

# Остановка конфликтующего сервиса
sudo systemctl stop <service>
```

### Проблемы с производительностью

```bash
# Проверка ресурсов
docker stats

# Увеличение лимитов в docker-compose.yml
services:
  go-app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

## Мониторинг производительности

### Метрики системы

```bash
# CPU и память
htop

# Сетевой трафик
iftop

# Disk I/O
iotop
```

### Метрики приложения

```bash
# Prometheus metrics
curl http://localhost:8080/metrics | grep vpn_

# Статистика сервиса
curl http://localhost:8080/stats | jq .
```

## Поддержка

При возникновении проблем:

1. Проверьте логи: `make logs`
2. Проверьте статус: `make status`
3. Проверьте health: `make test-health`
4. Откройте issue на GitHub

## Полезные команды

```bash
# Перезапуск сервиса
make restart

# Полная очистка и перезапуск
make clean && make up

# Просмотр активных пользователей
make list-users

# Backup перед важными операциями
make db-backup

# Проверка версий
docker --version
docker-compose --version
```

