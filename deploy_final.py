#!/usr/bin/env python3
import paramiko
import os
import tarfile

def deploy():
    print("🚀 Начинаем финальный деплой...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        print("✅ Подключение к серверу установлено")
        
        # Создаем архив проекта
        print("📦 Создаем архив проекта...")
        with tarfile.open('project.tar.gz', 'w:gz') as tar:
            for item in ['cmd', 'internal', 'migrations', 'pkg', 'go.mod', 'go.sum', 'docker-compose.yml', 'Dockerfile']:
                if os.path.exists(item):
                    tar.add(item)
        
        # Заливаем архив
        print("📤 Заливаем файлы на сервер...")
        sftp = ssh.open_sftp()
        sftp.put('project.tar.gz', '/tmp/project.tar.gz')
        
        # Устанавливаем Docker (правильно)
        print("🐳 Устанавливаем Docker...")
        docker_commands = [
            'apt remove -y docker docker-engine docker.io containerd runc || true',
            'apt update',
            'apt install -y apt-transport-https ca-certificates curl gnupg lsb-release',
            'curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg',
            'echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null',
            'apt update',
            'apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin',
            'docker --version',
            'docker compose version'
        ]
        
        for cmd in docker_commands:
            print(f"▶️  Выполняем: {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output.strip()}")
        
        # Запускаем приложение
        print("🚀 Запускаем приложение...")
        app_commands = [
            'mkdir -p /opt/auth-service',
            'cd /opt/auth-service && tar -xzf /tmp/project.tar.gz',
            'cd /opt/auth-service && docker compose down || true',
            'cd /opt/auth-service && docker compose up -d --build',
            'sleep 20'  # Ждем пока все запустится
        ]
        
        for cmd in app_commands:
            print(f"▶️  Выполняем: {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output.strip()}")
        
        # Проверяем статус
        print("🔍 Проверяем статус сервисов...")
        stdin, stdout, stderr = ssh.exec_command('cd /opt/auth-service && docker compose ps')
        status = stdout.read().decode()
        print(f"📊 Статус контейнеров:\n{status}")
        
        # Проверяем логи
        print("📋 Проверяем логи приложения...")
        stdin, stdout, stderr = ssh.exec_command('cd /opt/auth-service && docker compose logs auth-service')
        logs = stdout.read().decode()
        print(f"📝 Логи:\n{logs[-500:]}")  # Последние 500 символов
        
        print("🎉 Деплой завершен!")
        print("🌐 Приложение доступно по адресу: http://77.110.105.228:8080")
        print("🗄️  База данных доступна на: 77.110.105.228:5432")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()
        if os.path.exists('project.tar.gz'):
            os.remove('project.tar.gz')

if __name__ == "__main__":
    deploy()