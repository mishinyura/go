FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o auth-app ./auth/cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/auth-app .

EXPOSE 50051

CMD ["./auth-app"]