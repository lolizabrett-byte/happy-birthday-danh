package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite" // pure-Go driver, keeps CGO_ENABLED=0 builds working
)

// newSQLiteStore opens (or creates) a local SQLite file for self-hosting on a VPS.
// It reuses sqlStore: the schema and queries are plain SQLite, same as Turso (libSQL).
func newSQLiteStore(ctx context.Context, path string) (*sqlStore, error) {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "busy_timeout(5000)")
	db, err := sql.Open("sqlite", "file:"+path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // single writer: avoids SQLITE_BUSY under concurrent POSTs
	s := &sqlStore{db: db}
	if err := s.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}
