#!/usr/bin/env python3
import paramiko

def connect_to_server():
    print("🔗 Подключаемся к серверу...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        print("✅ Подключение установлено!")
        
        # Проверим логи приложения
        print("\n📝 Проверяем логи auth-service...")
        stdin, stdout, stderr = ssh.exec_command('cd /opt/auth-service && docker compose logs auth-service --tail=20')
        logs = stdout.read().decode()
        print(logs)
        
        # Проверим настройки SMTP
        print("\n🔧 Проверяем SMTP настройки...")
        stdin, stdout, stderr = ssh.exec_command('cd /opt/auth-service && cat .env | grep SMTP')
        smtp_settings = stdout.read().decode()
        print(smtp_settings)
        
        # Оставим сессию открытой для команд
        print("\n💻 Можешь вводить команды (для выхода введи 'exit'):")
        while True:
            cmd = input("server$ ")
            if cmd.lower() == 'exit':
                break
            stdin, stdout, stderr = ssh.exec_command(f'cd /opt/auth-service && {cmd}')
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(output)
            if error:
                print(f"Ошибка: {error}")
                
    except Exception as e:
        print(f"❌ Ошибка подключения: {e}")
    finally:
        ssh.close()

if __name__ == "__main__":
    connect_to_server()