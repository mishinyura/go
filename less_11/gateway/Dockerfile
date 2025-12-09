# gateway/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/gateway ./cmd/gateway

FROM alpine:3.18
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=builder /app/gateway .
EXPOSE 8080
USER nobody
CMD ["/app/gateway"]
