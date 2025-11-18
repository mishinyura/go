package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/you/monorepo/gateway/internal/api"
	"github.com/you/monorepo/ledger/internal/app"
)

func main() {
	ctx := context.Background()
	cfg := app.Config{
		DSN: os.Getenv("DATABASE_URL"),
	}
	svc, closeFn, err := app.NewServiceFactory(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init ledger service: %v", err)
	}
	defer closeFn()

	h := api.NewHandlers(svc) // конструктор, принимает service.Service

	mux := h.RegisterMux() // возвращает http.Handler или *http.ServeMux

	// graceful shutdown minimal
	srv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		log.Println("gateway listening :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutdown")
	_ = srv.Shutdown(ctx)
}
