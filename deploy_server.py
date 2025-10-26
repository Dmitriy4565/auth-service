#!/usr/bin/env python3
import paramiko
import os

def deploy_to_server():
    print("🚀 Деплоим на сервер...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        sftp = ssh.open_sftp()
        
        print("📤 Заливаем файлы на сервер...")
        
        def upload_file(local_path, remote_path):
            """Загружает файл на сервер"""
            try:
                sftp.put(local_path, remote_path)
                print(f"✅ {local_path} -> {remote_path}")
                return True
            except Exception as e:
                print(f"❌ Ошибка загрузки {local_path}: {e}")
                return False
        
        def upload_folder(local_folder, remote_folder):
            """Загружает папку на сервер"""
            for root, dirs, files in os.walk(local_folder):
                for file in files:
                    local_path = os.path.join(root, file)
                    relative_path = os.path.relpath(local_path, local_folder)
                    remote_path = f"{remote_folder}/{relative_path}"
                    remote_dir = os.path.dirname(remote_path)
                    
                    # Создаем папки на сервере
                    try:
                        sftp.stat(remote_dir)
                    except:
                        ssh.exec_command(f"mkdir -p '{remote_dir}'")
                    
                    upload_file(local_path, remote_path)
        
        # Заливаем папки
        folders = ['cmd', 'internal', 'migrations', 'pkg']
        for folder in folders:
            if os.path.exists(folder):
                print(f"📁 Заливаем папку: {folder}")
                upload_folder(folder, f"/opt/auth-service/{folder}")
        
        # Заливаем отдельные файлы
        files = ['go.mod', 'go.sum', 'docker-compose.yml', 'Dockerfile', '.env']
        for file in files:
            if os.path.exists(file):
                print(f"📄 Заливаем файл: {file}")
                upload_file(file, f"/opt/auth-service/{file}")
        
        print("🐳 Перезапускаем сервис на сервере...")
        
        commands = [
            'cd /opt/auth-service',
            'echo "🛑 Останавливаем старые контейнеры..."',
            'docker compose down',
            'echo "🔨 Собираем новый образ..."',
            'docker compose up -d --build',
            'echo "⏳ Ждем 20 секунд..."',
            'sleep 20',
            'echo "📊 Проверяем статус контейнеров:"',
            'docker compose ps',
            'echo "📝 Логи приложения:"',
            'docker compose logs auth-service --tail=15'
        ]
        
        for cmd in commands:
            print(f"▶️  Выполняем: {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd, timeout=60)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output}")
            if error and "WARNING" not in error:
                print(f"⚠️  {error}")
        
        print("🎉 Деплой завершен!")
        print("🌐 API доступно: http://77.110.105.228:8080")
        print("📧 Письма отправляются через Яндекс SMTP")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()

if __name__ == "__main__":
    deploy_to_server()