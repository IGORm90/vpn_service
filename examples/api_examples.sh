#!/bin/bash

# Примеры использования API VPN сервиса

BASE_URL="http://localhost:8080"

echo "=== VPN Service API Examples ==="
echo ""

# Health Check
echo "1. Health Check"
curl -s "$BASE_URL/health" | jq .
echo ""

# Create User
echo "2. Create User"
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "securepass123",
    "traffic_limit": 10737418240,
    "expires_at": "2025-12-31T23:59:59Z"
  }')
echo "$USER_RESPONSE" | jq .
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id')
echo "Created user with ID: $USER_ID"
echo ""

# List Users
echo "3. List All Users"
curl -s "$BASE_URL/api/users" | jq .
echo ""

# Get User
echo "4. Get User Details"
curl -s "$BASE_URL/api/users/$USER_ID" | jq .
echo ""

# Get User Config
echo "5. Get User Configuration"
curl -s "$BASE_URL/api/users/$USER_ID/config" | jq .
echo ""

# Update User
echo "6. Update User (increase traffic limit)"
curl -s -X PATCH "$BASE_URL/api/users/$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "traffic_limit": 21474836480
  }' | jq .
echo ""

# Reset Traffic
echo "7. Reset User Traffic"
curl -s -X POST "$BASE_URL/api/users/$USER_ID/reset-traffic" | jq .
echo ""

# Get Stats
echo "8. Get Service Stats"
curl -s "$BASE_URL/stats" | jq .
echo ""

# Prometheus Metrics
echo "9. Prometheus Metrics (sample)"
curl -s "$BASE_URL/metrics" | grep "^vpn_" | head -n 10
echo ""

# Delete User (uncomment to test)
# echo "10. Delete User"
# curl -s -X DELETE "$BASE_URL/api/users/$USER_ID"
# echo ""

echo "=== Examples completed ==="

