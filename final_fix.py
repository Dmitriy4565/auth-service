#!/usr/bin/env python3
import paramiko

def fix_and_restart():
    print("🔧 Исправляем и запускаем...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        
        # Одна большая команда чтобы избежать проблем с cd
        cmd = """
        cd /opt/auth-service && \
        echo "🐳 Проверяем версию Docker..." && \
        docker --version && \
        docker compose version && \
        echo "🛑 Останавливаем старые контейнеры..." && \
        docker compose down && \
        echo "🔨 Собираем новый образ с Яндекс SMTP..." && \
        docker compose up -d --build && \
        echo "⏳ Ждем 20 секунд..." && \
        sleep 20 && \
        echo "📊 Статус контейнеров:" && \
        docker compose ps && \
        echo "📝 Логи приложения:" && \
        docker compose logs auth-service --tail=10
        """
        
        print("▶️  Выполняем команды...")
        stdin, stdout, stderr = ssh.exec_command(cmd, timeout=180)
        output = stdout.read().decode()
        error = stderr.read().decode()
        
        print("📋 Вывод:")
        print(output)
        if error:
            print("⚠️  Ошибки:")
            print(error)
        
        print("🎉 Готово! Проверяй: http://77.110.105.228:8080")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()

if __name__ == "__main__":
    fix_and_restart()