FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o auth-service ./cmd/server

EXPOSE 8080

CMD ["./auth-service"]