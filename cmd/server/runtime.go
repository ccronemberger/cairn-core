package main

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencairn-org/cairn-core/internal/config"
)

type appRuntime struct {
	ctx    context.Context
	pool   *pgxpool.Pool
	cfg    *config.Config
	logger *slog.Logger
}

func newRuntime(baseCtx context.Context) (*appRuntime, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := newLogger(cfg.Log)
	slog.SetDefault(logger)

	// ------------------------------------------------------------------
	// Database
	// ------------------------------------------------------------------

	pool, err := serveOpenDatabase(baseCtx, cfg, logger)
	if err != nil {
		return nil, err
	}
	// defer pool.Close() - moved to the close method

	// ------------------------------------------------------------------
	// Schema version check / auto-migrate
	// ------------------------------------------------------------------

	if err := ensureSchema(baseCtx, logger, pool, cfg.Database.AutoMigrate); err != nil {
		return nil, err
	}

	return &appRuntime{
		ctx:    baseCtx,
		pool:   pool,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (r *appRuntime) close() {
	r.pool.Close()
}
