#!/usr/bin/env python3
"""
Пример Python клиента для работы с VPN Service API
"""

import requests
import json
from datetime import datetime, timedelta

class VPNServiceClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
    
    def health_check(self):
        """Проверка здоровья сервиса"""
        response = self.session.get(f"{self.base_url}/health")
        return response.json()
    
    def get_stats(self):
        """Получить статистику сервиса"""
        response = self.session.get(f"{self.base_url}/stats")
        return response.json()
    
    def create_user(self, username, password, traffic_limit=0, days=30):
        """Создать пользователя"""
        expires_at = (datetime.utcnow() + timedelta(days=days)).isoformat() + "Z"
        
        data = {
            "username": username,
            "password": password,
            "traffic_limit": traffic_limit,
            "expires_at": expires_at
        }
        
        response = self.session.post(
            f"{self.base_url}/api/users",
            json=data
        )
        return response.json()
    
    def list_users(self, active_only=False):
        """Получить список пользователей"""
        params = {"active": "true"} if active_only else {}
        response = self.session.get(
            f"{self.base_url}/api/users",
            params=params
        )
        return response.json()
    
    def get_user(self, user_id):
        """Получить информацию о пользователе"""
        response = self.session.get(f"{self.base_url}/api/users/{user_id}")
        return response.json()
    
    def update_user(self, user_id, **kwargs):
        """Обновить пользователя"""
        response = self.session.patch(
            f"{self.base_url}/api/users/{user_id}",
            json=kwargs
        )
        return response.json()
    
    def delete_user(self, user_id):
        """Удалить пользователя"""
        response = self.session.delete(f"{self.base_url}/api/users/{user_id}")
        return response.status_code == 204
    
    def get_user_config(self, user_id):
        """Получить конфигурацию для клиента"""
        response = self.session.get(
            f"{self.base_url}/api/users/{user_id}/config"
        )
        return response.json()
    
    def reset_traffic(self, user_id):
        """Сбросить трафик пользователя"""
        response = self.session.post(
            f"{self.base_url}/api/users/{user_id}/reset-traffic"
        )
        return response.json()


def main():
    # Создаем клиент
    client = VPNServiceClient()
    
    # Проверяем здоровье
    print("Health Check:", json.dumps(client.health_check(), indent=2))
    
    # Создаем пользователя
    print("\nCreating user...")
    user = client.create_user(
        username="test_python_user",
        password="python123",
        traffic_limit=10 * 1024**3,  # 10GB
        days=30
    )
    print(json.dumps(user, indent=2))
    
    if user.get("success"):
        user_id = user["data"]["id"]
        
        # Получаем конфигурацию
        print(f"\nGetting config for user {user_id}...")
        config = client.get_user_config(user_id)
        print("VLESS URI:", config["data"]["uri"])
        
        # Обновляем лимит
        print(f"\nUpdating user {user_id}...")
        updated = client.update_user(user_id, traffic_limit=20 * 1024**3)
        print(json.dumps(updated, indent=2))
        
        # Получаем статистику
        print("\nService Stats:")
        stats = client.get_stats()
        print(json.dumps(stats, indent=2))
        
        # Удаляем пользователя (раскомментируйте при необходимости)
        # print(f"\nDeleting user {user_id}...")
        # client.delete_user(user_id)


if __name__ == "__main__":
    main()

