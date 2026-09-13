package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"sap_segmentation/internal/config"
	"sap_segmentation/internal/importer"
	"sap_segmentation/internal/logx"
)

func main() {
	if err := logx.Init(); err != nil {
		log.Fatalf("init logger: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logx.CleanupOld(cfg.LogCleanupMaxAge)

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := importer.Run(ctx, cfg, db); err != nil {
		log.Printf("import error: %v", err)
		os.Exit(1)
	}
}
