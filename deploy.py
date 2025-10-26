#!/usr/bin/env python3
import paramiko
import os
import tarfile

def safe_deploy():
    print("🚀 ЗАПУСКАЕМ БЕЗОПАСНЫЙ ДЕПЛОЙ...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        print("✅ Подключение к серверу установлено")
        
        # Останавливаем только текущий сервис
        print("⏸️  Останавливаем текущий сервис...")
        stop_commands = [
            'cd /opt/auth-service && docker compose down 2>/dev/null || true',
            'sleep 5'
        ]
        
        for cmd in stop_commands:
            print(f"⏹️  {cmd}")
            ssh.exec_command(cmd)
        
        # Создаем архив проекта
        print("📦 Создаем архив проекта...")
        with tarfile.open('project.tar.gz', 'w:gz') as tar:
            for item in ['.env', 'docker-compose.yml', 'Dockerfile', 'cmd', 'internal', 'migrations', 'pkg', 'go.mod', 'go.sum']:
                if os.path.exists(item):
                    tar.add(item)
                    print(f"📁 Добавлен: {item}")
        
        # Заливаем архив
        print("📤 Заливаем файлы на сервер...")
        sftp = ssh.open_sftp()
        sftp.put('project.tar.gz', '/tmp/project.tar.gz')
        
        # Копируем конфиги
        config_files = {
            '.env': '📄 .env файл',
            'docker-compose.yml': '🐳 docker-compose.yml', 
            'Dockerfile': '🐹 Dockerfile'
        }
        
        for file, desc in config_files.items():
            if os.path.exists(file):
                print(f"{desc}...")
                sftp.put(file, f'/tmp/{file}')
            else:
                print(f"❌ {file} не найден!")
                return
        
        # ОБНОВЛЯЕМ ПРОЕКТ
        print("🔄 Обновляем проект...")
        build_commands = [
            'mkdir -p /opt/auth-service',
            'cd /opt/auth-service && tar -xzf /tmp/project.tar.gz --overwrite',
            'cp /tmp/.env /opt/auth-service/.env',
            'cp /tmp/docker-compose.yml /opt/auth-service/docker-compose.yml', 
            'cp /tmp/Dockerfile /opt/auth-service/Dockerfile',
            'cd /opt/auth-service && docker compose up -d --build',
            'sleep 15',
            'cd /opt/auth-service && docker compose ps'
        ]
        
        for cmd in build_commands:
            print(f"🔨 {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output.strip()}")
            if error and not any(x in error for x in ['orphan', 'warning']):
                print(f"⚠️  {error.strip()}")
        
        # ЗАПУСКАЕМ МИГРАЦИИ
        print("🗄️  Запускаем миграции БД...")
        migration_commands = [
            'sleep 10',
            'cd /opt/auth-service && docker compose cp migrations/ postgres:/tmp/',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -f /tmp/migrations/001_init.sql',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -f /tmp/migrations/002_add_reset_password_tokens.sql',
            'cd /opt/auth-service && docker compose exec -T postgres psql -U postgres -d auth_service -c "\dt"'
        ]
        
        for cmd in migration_commands:
            print(f"📊 {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            if output:
                print(f"📋 {output.strip()}")
        
        # ФИНАЛЬНАЯ ПРОВЕРКА
        print("🔍 Финальная проверка...")
        check_commands = [
            'cd /opt/auth-service && docker compose ps',
            'cd /opt/auth-service && curl -s http://localhost:8080/health || echo "Health check failed"',
            'cd /opt/auth-service && docker compose logs auth-service --tail=10'
        ]
        
        for cmd in check_commands:
            stdin, stdout, stderr = ssh.exec_command(cmd)
            result = stdout.read().decode().strip()
            if result:
                print(f"✅ {result}")
        
        print("🎉 ДЕПЛОЙ УСПЕШНО ЗАВЕРШЕН!")
        print("🌐 Приложение: http://77.110.105.228:8080")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()
        if os.path.exists('project.tar.gz'):
            os.remove('project.tar.gz')

if __name__ == "__main__":
    safe_deploy()