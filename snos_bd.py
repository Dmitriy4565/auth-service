#!/usr/bin/env python3
import paramiko

def snos_bd():
    print("🗑️  ЗАПУСКАЕМ ОЧИСТКУ БАЗ ДАННЫХ...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        print("✅ Подключение к серверу установлено")
        
        # Команды для очистки БД
        cleanup_commands = [
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -c "DROP DATABASE IF EXISTS auth_service;"',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -c "CREATE DATABASE auth_service;"',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -f /tmp/migrations/001_init.sql',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -f /tmp/migrations/002_add_reset_password_tokens.sql',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -c "\dt"'
        ]
        
        for cmd in cleanup_commands:
            print(f"🧹 {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output.strip()}")
            if error:
                print(f"⚠️  {error.strip()}")
        
        print("✅ БАЗЫ ДАННЫХ ОЧИЩЕНЫ И ПЕРЕСОЗДАНЫ!")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()

if __name__ == "__main__":
    snos_bd()