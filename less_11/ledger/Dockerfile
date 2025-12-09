# ledger/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /app/ledger ./cmd/ledger

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/ledger .
EXPOSE 9090
USER nobody
CMD ["/app/ledger"]
