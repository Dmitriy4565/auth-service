version: '3.8'

networks:
  microservices-net:
    driver: bridge

services:
  auth-service:
    build: 
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=password
      - DB_NAME=auth_service
      - PORT=8080
      - GIN_MODE=debug
      - JWT_SECRET=WgHx8L3pF2qR9tY1vK6zM0nB7cJ4dA5sX8eP1rT3yU6iO9wQ2fS5hV7kZ0lC4jN
      - CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:8081,http://192.168.31.173:3000,http://192.168.191.226:3000
      - CORS_ALLOW_CREDENTIALS=true
      - ACCESS_TOKEN_EXPIRE_MINUTES=15
      - REFRESH_TOKEN_EXPIRE_DAYS=7
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=prihodin816@gmail.com
      - SMTP_PASSWORD=pdka bfpm zbct ylbu
      - SMTP_FROM=prihodin816@gmail.com
      - CLIENT_URL=http://localhost:3000
    restart: unless-stopped
    networks:
      - microservices-net
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: auth_service
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped
    networks:
      - microservices-net
    ports:
      - "5432:5432"

volumes:
  postgres_data: