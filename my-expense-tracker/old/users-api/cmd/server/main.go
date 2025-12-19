package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	_ "gitlab.com/education/users-api/internal/docs"
	"gitlab.com/education/users-api/internal/grpcserver"
	"gitlab.com/education/users-api/internal/handler"
	pb "gitlab.com/education/users-api/internal/pb/budget/v1"
	"gitlab.com/education/users-api/internal/repository"
	"gitlab.com/education/users-api/internal/service"
	"gitlab.com/education/users-api/internal/ui"
)

// @title           Users API
// @version         1.0
// @description     Пример API пользователей на Gin
// @BasePath        /api/v1
// @schemes         http
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := gin.Default()

	userRepo := repository.NewInMemoryUserRepository(nil)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	sheetsClient, err := service.NewSheetsClient(ctx, credentialsPath)
	if err != nil {
		log.Fatalf("init sheets client: %v", err)
	}

	sheetID := os.Getenv("DEMO_SPREADSHEET_ID")
	if sheetID == "" {
		log.Fatal("environment variable DEMO_SPREADSHEET_ID is required")
	}
	sheetName := os.Getenv("DEMO_SHEET_NAME")
	if sheetName == "" {
		sheetName = "Report"
	}

	// настройки “по умолчанию” позволяют дергать download без параметров
	demoCfg := service.DemoSheetConfig{
		SpreadsheetID: sheetID,
		SheetName:     sheetName,
	}

	demoSvc := service.NewSheetDemoService(sheetsClient, demoCfg)
	demoHandler := handler.NewDemoSheetHandler(demoSvc)

	api := r.Group("/api/v1")
	userHandler.Register(api)
	demoHandler.Register(api)

	// Swagger UI (gin-swagger): /swagger/index.html и /swagger/doc.json
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ReDoc (через embed)
	r.GET("/redoc", func(c *gin.Context) {
		b, err := ui.ReDocHTML()
		if err != nil {
			c.String(http.StatusInternalServerError, "redoc template error: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", b)
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	grpcAddr := ":" + grpcPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen gRPC: %v", err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterBudgetServiceServer(grpcSrv, grpcserver.NewBudgetServer(demoSvc))

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("HTTP server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP shutdown error: %v", err)
		}
		grpcSrv.GracefulStop()
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server error: %v", err)
	}
}
