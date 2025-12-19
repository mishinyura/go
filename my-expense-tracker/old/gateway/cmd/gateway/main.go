package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gitlab.com/education/gateway/internal/config"
	"gitlab.com/education/gateway/internal/handler"
	budgetv1 "gitlab.com/education/gateway/internal/pb/budget/v1"
	"gitlab.com/education/gateway/internal/server/httpserver"
	"gitlab.com/education/gateway/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	log.Printf("Connecting to gRPC backend at %s", cfg.GRPC.Address)
	conn, err := dialGRPC(ctx, cfg.GRPC.Address)
	if err != nil {
		log.Fatalf("failed to connect to gRPC backend: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to gRPC backend successfully")

	budgetService := service.NewBudgetGatewayService(budgetv1.NewBudgetServiceClient(conn))
	budgetHandler := handler.NewBudgetHandler(budgetService)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	api := engine.Group("/api")
	budgetHandler.Register(api)

	server := httpserver.New(cfg.HTTP, engine)

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTP.Address)
		if err := server.Start(); err != nil {
			log.Fatalf("http server stopped: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func dialGRPC(ctx context.Context, address string) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
