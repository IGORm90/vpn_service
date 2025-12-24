#!/bin/bash

# Скрипт для создания тестовых пользователей

BASE_URL="http://localhost:8080"

echo "Creating test users..."

# User 1: Free tier (5GB, 30 days)
curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "free_user",
    "password": "password123",
    "traffic_limit": 5368709120,
    "expires_at": "'$(date -u -v+30d +%Y-%m-%dT%H:%M:%SZ)'"
  }' | jq .

# User 2: Basic tier (50GB, 90 days)
curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "basic_user",
    "password": "password123",
    "traffic_limit": 53687091200,
    "expires_at": "'$(date -u -v+90d +%Y-%m-%dT%H:%M:%SZ)'"
  }' | jq .

# User 3: Premium tier (unlimited, 365 days)
curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "premium_user",
    "password": "password123",
    "traffic_limit": 0,
    "expires_at": "'$(date -u -v+365d +%Y-%m-%dT%H:%M:%SZ)'"
  }' | jq .

echo ""
echo "Test users created successfully!"

