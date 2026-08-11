package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	"github.com/xdars/grpc-tasks/internal/db"
	grpcserver "github.com/xdars/grpc-tasks/internal/grpc"
	"github.com/xdars/grpc-tasks/internal/grpc/auth"
	"github.com/xdars/grpc-tasks/internal/grpc/payments"
	"github.com/xdars/grpc-tasks/internal/grpc/tasks"
	httpserver "github.com/xdars/grpc-tasks/internal/http"
	"github.com/xdars/grpc-tasks/internal/store"
	wa "github.com/xdars/grpc-tasks/internal/webauthn"
	"github.com/xdars/grpc-tasks/pkg/config"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbURL string) error {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return fmt.Errorf("migrations init: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrations up: %w", err)
	}
	log.Println("migrations applied")
	return nil
}

func Run() error {
	cfg := config.Load()

	pool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	userRepo := db.NewUserRepository(pool)
	tasksRepo := db.NewTaskRepository(pool)
	waRepo := db.NewWebAuthnRepository(pool)
	paymentsRepo := db.NewPaymentRepository(pool)

	s := store.NewStore()

	authSvc := auth.NewAuthService(userRepo)
	taskSvc := tasks.NewTaskService(tasksRepo, s)
	waSvc, err := wa.NewService(userRepo, waRepo)
	paymentSvc := payments.NewPaymentService(paymentsRepo)

	if err != nil {
		return fmt.Errorf("webauthn init: %w", err)
	}

	grpcSrv := grpcserver.New(authSvc, taskSvc, paymentSvc)
	lis, err := grpcSrv.Listen(cfg.GRPCPort)

	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	httpSrv := httpserver.New(authSvc, taskSvc, s, waSvc)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: httpSrv.Routes(),
	}

	go func() {
		log.Printf("gRPC listening on :%d", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC error: %v", err)
		}
	}()

	go func() {
		log.Printf("HTTP listening on :%d", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	grpcSrv.GracefulStop()
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	return nil
}
