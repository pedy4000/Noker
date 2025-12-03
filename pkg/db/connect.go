package db

import (
	"database/sql"
	"time"

	"github.com/pedy4000/noker/pkg/config"
	"github.com/pedy4000/noker/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

func Connect(dsn string, cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	logger.Info("Connected to PostgreSQL")
	return db, nil
}
