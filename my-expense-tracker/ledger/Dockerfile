FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# Собираем сервис Ledger
RUN go build -o ledger-app ./ledger/cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/ledger-app .

# Порт леджера
EXPOSE 50052

CMD ["./ledger-app"]