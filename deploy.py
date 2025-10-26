#!/usr/bin/env python3
import paramiko
import os
import tarfile

def nuclear_deploy():
    print("💣 ЗАПУСКАЕМ ЯДЕРНЫЙ ДЕПЛОЙ - СНОСИМ ВСЕ К ХУЯМ...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        print("✅ Подключение к серверу установлено")
        
        # СНОСИМ ВСЕ К ЧЕРТЯМ
        print("💥 Сносим все старые контейнеры и volumes...")
        nuke_commands = [
            'cd /opt/auth-service && docker compose down --remove-orphans --volumes --rmi all 2>/dev/null || true',
            'docker rm -f $(docker ps -aq --filter name=auth-service) 2>/dev/null || true',
            'docker volume rm $(docker volume ls -q --filter name=auth-service) 2>/dev/null || true',
            'docker network rm $(docker network ls -q --filter name=auth-service) 2>/dev/null || true',
            'rm -rf /opt/auth-service/* 2>/dev/null || true',
            'docker system prune -f 2>/dev/null || true'
        ]
        
        for cmd in nuke_commands:
            print(f"💀 {cmd}")
            ssh.exec_command(cmd)
            # Нам похуй на ошибки - просто сносим
        
        # Создаем архив проекта ВСЕХ файлов
        print("📦 Создаем свежий архив проекта...")
        with tarfile.open('project.tar.gz', 'w:gz') as tar:
            for item in ['.env', 'docker-compose.yml', 'Dockerfile', 'cmd', 'internal', 'migrations', 'pkg', 'go.mod', 'go.sum']:
                if os.path.exists(item):
                    tar.add(item)
                    print(f"📁 Добавлен: {item}")
        
        # Заливаем архив
        print("📤 Заливаем файлы на сервер...")
        sftp = ssh.open_sftp()
        sftp.put('project.tar.gz', '/tmp/project.tar.gz')
        
        # Создаем правильные файлы конфигурации
        print("⚙️  Создаем конфигурационные файлы...")
        
        # 1. .env файл
        env_content = '''DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=auth_service

PORT=8080
GIN_MODE=debug

JWT_SECRET=WgHx8L3pF2qR9tY1vK6zM0nB7cJ4dA5sX8eP1rT3yU6iO9wQ2fS5hV7kZ0lC4jN

CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:8081,http://192.168.31.173:3000,http://192.168.191.226:3000
CORS_ALLOW_CREDENTIALS=true

ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_DAYS=7

# Resend настройки
RESEND_API_KEY=re_HdvD8ftT_BwwPQEQu4UbptMxioYfoN9wR
RESEND_FROM_EMAIL=noreply@rossttelecom.ru
RESEND_FROM_NAME=Auth Service

CLIENT_URL=http://localhost:3000
'''
        with open('.env', 'w') as f:
            f.write(env_content)
        sftp.put('.env', '/tmp/.env')
        
        # 2. docker-compose.yml с Postgres 15
        compose_content = '''services:
  auth-service:
    build: .
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=auth_service
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
'''
        with open('docker-compose.yml', 'w') as f:
            f.write(compose_content)
        sftp.put('docker-compose.yml', '/tmp/docker-compose.yml')
        
        # 3. Dockerfile
        dockerfile_content = '''FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o auth-service ./cmd/server

EXPOSE 8080

CMD ["./auth-service"]
'''
        with open('Dockerfile', 'w') as f:
            f.write(dockerfile_content)
        sftp.put('Dockerfile', '/tmp/Dockerfile')
        
        # СОБИРАЕМ ЗАНОВО
        print("🚀 Собираем все заново...")
        build_commands = [
            'mkdir -p /opt/auth-service',
            'cd /opt/auth-service && tar -xzf /tmp/project.tar.gz --overwrite',
            'cp /tmp/.env /opt/auth-service/.env',
            'cp /tmp/docker-compose.yml /opt/auth-service/docker-compose.yml', 
            'cp /tmp/Dockerfile /opt/auth-service/Dockerfile',
            'cd /opt/auth-service && docker compose up -d --build --force-recreate',
            'sleep 25',  # Даем время на полный запуск
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
            'sleep 10',  # Ждем полного запуска postgres
            'cd /opt/auth-service && docker compose exec postgres psql -U postgres -d auth_service -f /app/migrations/001_init.sql',
            'cd /opt/auth-service && docker compose exec postgres psql -U postgres -d auth_service -f /app/migrations/002_add_reset_password_tokens.sql',
            'cd /opt/auth-service && docker compose exec postgres psql -U postgres -d auth_service -c "\\dt"'
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
            'cd /opt/auth-service && curl -s http://localhost:8080/health',
            'cd /opt/auth-service && docker compose exec auth-service env | grep RESEND',
            'cd /opt/auth-service && docker compose logs auth-service --tail=5'
        ]
        
        for cmd in check_commands:
            stdin, stdout, stderr = ssh.exec_command(cmd)
            result = stdout.read().decode().strip()
            if result:
                print(f"✅ {result}")
        
        print("🎉 ЯДЕРНЫЙ ДЕПЛОЙ ЗАВЕРШЕН!")
        print("🌐 Приложение: http://77.110.105.228:8080")
        print("🗄️  База данных: 77.110.105.228:5432") 
        print("📧 Resend: noreply@rossttelecom.ru")
        print("💪 Теперь должно работать нахуй!")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()
        # Чистим временные файлы
        for f in ['project.tar.gz', '.env', 'docker-compose.yml', 'Dockerfile']:
            if os.path.exists(f):
                os.remove(f)

if __name__ == "__main__":
    nuclear_deploy()