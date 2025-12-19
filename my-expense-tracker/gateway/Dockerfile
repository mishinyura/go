FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# Собираем Gateway
RUN go build -o gateway-app ./gateway/cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/gateway-app .

# Порт Gateway (HTTP)
EXPOSE 8080

CMD ["./gateway-app"]