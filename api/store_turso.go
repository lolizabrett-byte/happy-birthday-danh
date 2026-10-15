package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

// storeTimeout bounds every query so a stalled Turso cannot hang a request.
const storeTimeout = 5 * time.Second

// Fixed-width UTC layout so TEXT columns sort chronologically (RFC3339Nano trims zeros and does not).
const dbTimeLayout = "2006-01-02T15:04:05.000Z"

const schema = `
CREATE TABLE IF NOT EXISTS wishes (
	id TEXT PRIMARY KEY, name TEXT NOT NULL, message TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS scores (
	id TEXT PRIMARY KEY, name TEXT NOT NULL, score INTEGER NOT NULL, created_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_scores_rank ON scores (score DESC, created_at);`

// tursoStore talks to Turso over the pure-Go libsql client (no CGO).
type tursoStore struct{ db *sql.DB }

// newTursoStore connects and creates tables. The driver maps libsql:// to https (stateless HTTP requests,
// no websocket), so the URL from the Turso dashboard is used as-is.
func newTursoStore(ctx context.Context, url, token string) (*tursoStore, error) {
	var opts []libsql.Option
	if token != "" {
		opts = append(opts, libsql.WithAuthToken(token))
	}
	conn, err := libsql.NewConnector(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("turso connector: %w", err)
	}
	db := sql.OpenDB(conn)
	db.SetConnMaxLifetime(5 * time.Minute)
	s := &tursoStore{db: db}
	if err := s.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *tursoStore) init(ctx context.Context) error {
	for _, stmt := range strings.Split(schema, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}
	return nil
}

func (s *tursoStore) Close() error { return s.db.Close() }

func (s *tursoStore) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	var one int
	return s.db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
}

func (s *tursoStore) AddWish(ctx context.Context, w Wish) error {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO wishes (id, name, message, created_at) VALUES (?, ?, ?, ?)",
		w.ID, w.Name, w.Message, w.CreatedAt.UTC().Format(dbTimeLayout))
	return err
}

func (s *tursoStore) ListWishes(ctx context.Context, limit int) ([]Wish, error) {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, message, created_at FROM wishes ORDER BY created_at DESC, rowid DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Wish{}
	for rows.Next() {
		var w Wish
		var at string
		if err := rows.Scan(&w.ID, &w.Name, &w.Message, &at); err != nil {
			return nil, err
		}
		if w.CreatedAt, err = time.Parse(dbTimeLayout, at); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *tursoStore) AddScore(ctx context.Context, sc Score) error {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO scores (id, name, score, created_at) VALUES (?, ?, ?, ?)",
		sc.ID, sc.Name, sc.Score, sc.CreatedAt.UTC().Format(dbTimeLayout))
	return err
}

func (s *tursoStore) TopScores(ctx context.Context, limit int) ([]Score, error) {
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, score, created_at FROM scores ORDER BY score DESC, created_at ASC, rowid ASC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Score{}
	for rows.Next() {
		var sc Score
		var at string
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.Score, &at); err != nil {
			return nil, err
		}
		if sc.CreatedAt, err = time.Parse(dbTimeLayout, at); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}
