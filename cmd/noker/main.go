package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pedy4000/noker/internal/api"
	"github.com/pedy4000/noker/internal/queue"
	"github.com/pedy4000/noker/internal/repository"
	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/db"
	"github.com/pedy4000/noker/pkg/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}
	fmt.Println(cfg.Database.URL)

	// Initialize logger
	logger.Init(cfg.Env == "development")
	logger.Info("Running on", cfg.Env, "Mode")

	// Connect to database
	dbConn, err := db.Connect(cfg.Database.URL, cfg)
	if err != nil {
		log.Fatal("database connection failed:", err)
	}

	queries := repository.New(dbConn)

	// Initialize queue processor
	processor := queue.NewProcessor(queries, cfg)

	// Create handler (API needs processor for enqueue)
	handler := api.NewHandler(queries, processor)

	// Create router
	router := api.NewRouter(handler, cfg)

	// Start HTTP server (only if enabled — for worker-only mode)
	var srv *http.Server
	if cfg.Server.Enabled {
		addr := ":" + cfg.Server.Port
		srv = &http.Server{
			Addr:    addr,
			Handler: router,
		}

		go func() {
			logger.Info("API server starting on " + addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("API server failed:", err)
			}
		}()
	} else {
		logger.Info("API server disabled — running in worker-only mode")
	}

	// Start background worker (only if AI is enabled)
	if cfg.AI.Enabled {
		go func() {
			logger.Info("Starting background AI worker...")
			processor.Start()
		}()
	} else {
		logger.Info("AI worker disabled via config")
	}

	// Graceful shutdown
	waitForShutdown(srv, dbConn)
}

func waitForShutdown(srv *http.Server, dbConn *sql.DB) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Shutdown signal received...")

	// Shutdown HTTP server
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("HTTP server forced to shutdown:", err)
		} else {
			logger.Info("HTTP server stopped gracefully")
		}
	}
	dbConn.Close()

	// Give worker time to finish current jobs
	logger.Info("Noker stopped")
}
