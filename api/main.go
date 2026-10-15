package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, closeStore, err := newStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeStore()

	e := newServer(store, cfg)
	e.Server.ReadHeaderTimeout = 5 * time.Second
	e.Server.ReadTimeout = 10 * time.Second
	e.Server.WriteTimeout = 15 * time.Second
	e.Server.IdleTimeout = 60 * time.Second
	errCh := make(chan error, 1)
	addr := cfg.Host + ":" + cfg.Port
	go func() { errCh <- e.Start(addr) }()
	slog.Info("listening", "addr", addr)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
	}
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return e.Shutdown(shutdownCtx)
}

// newStore picks turso when TURSO_DATABASE_URL is set, a local SQLite file when SQLITE_PATH is set,
// otherwise the in-memory dev store (refused on Render, where it would lose every wish on restart).
func newStore(ctx context.Context, cfg Config) (Store, func(), error) {
	if cfg.TursoURL == "" && cfg.SQLitePath != "" {
		initCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		ss, err := newSQLiteStore(initCtx, cfg.SQLitePath)
		if err != nil {
			return nil, nil, err
		}
		slog.Info("using sqlite store", "path", cfg.SQLitePath)
		return ss, func() { _ = ss.Close() }, nil
	}
	if cfg.TursoURL == "" {
		if cfg.OnRender {
			return nil, nil, errors.New("TURSO_DATABASE_URL is required when running on Render")
		}
		slog.Warn("TURSO_DATABASE_URL empty: using in-memory store (data lost on restart)")
		return newMemoryStore(), func() {}, nil
	}
	initCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ts, err := newTursoStore(initCtx, cfg.TursoURL, cfg.TursoToken)
	if err != nil {
		return nil, nil, err
	}
	return ts, func() { _ = ts.Close() }, nil
}
