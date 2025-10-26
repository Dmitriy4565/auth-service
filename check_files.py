#!/usr/bin/env python3
import paramiko

def check_server_files():
    print("🔍 Проверяем файлы на сервере...")
    
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        ssh.connect('77.110.105.228', username='root', password='WFdYPuq0Dyef')
        
        commands = [
            'cd /opt/auth-service && pwd',
            'cd /opt/auth-service && ls -la',
            'cd /opt/auth-service && ls -la migrations/',
            'cd /opt/auth-service && cat docker-compose.yml | head -10'
        ]
        
        for cmd in commands:
            print(f"▶️  {cmd}")
            stdin, stdout, stderr = ssh.exec_command(cmd)
            output = stdout.read().decode()
            error = stderr.read().decode()
            if output:
                print(f"📋 {output}")
            if error:
                print(f"⚠️  {error}")
        
    except Exception as e:
        print(f"❌ Ошибка: {e}")
    finally:
        ssh.close()

if __name__ == "__main__":
    check_server_files()