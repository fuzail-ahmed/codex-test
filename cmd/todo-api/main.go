package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/fuzail-ahmed/codex-test/internal/config"
	"github.com/fuzail-ahmed/codex-test/internal/repository/postgres"
	"github.com/fuzail-ahmed/codex-test/internal/server"
	"github.com/fuzail-ahmed/codex-test/internal/service"
	grpcserver "github.com/fuzail-ahmed/codex-test/internal/transport/grpc"
	httptransport "github.com/fuzail-ahmed/codex-test/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	repo := postgres.New(db)
	svc := service.New(repo, cfg.WorkerCount)

	mux := http.NewServeMux()
	handler := httptransport.NewHandler(svc)
	handler.Register(mux)

	httpServer := server.NewHTTP(cfg.HTTPAddr, mux)

	grpcSrv, grpcLis, err := grpcserver.ListenAndServe(cfg.GRPCAddr, svc)
	if err != nil {
		log.Fatalf("grpc listen error: %v", err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		grpcSrv.GracefulStop()
	}()

	go func() {
		log.Printf("grpc listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	log.Printf("http listening on %s", cfg.HTTPAddr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server error: %v", err)
	}
}
