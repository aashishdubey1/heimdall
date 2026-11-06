package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aashishdubey1/heimdall/internal/config"
	"github.com/aashishdubey1/heimdall/internal/database"
	"github.com/aashishdubey1/heimdall/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config %v:", err)
	}

	logger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("failed to load logger %v", err)
	}

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("database connection established")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	logger.Info("starting http server", "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		logger.Error("failed to start server ", "error", err)
		os.Exit(1)
	}

}
