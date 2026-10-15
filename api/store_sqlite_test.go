package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func openTestSQLite(t *testing.T, path string) *sqlStore {
	t.Helper()
	s, err := newSQLiteStore(context.Background(), path)
	if err != nil {
		t.Fatalf("newSQLiteStore: %v", err)
	}
	return s
}

func TestSQLiteStoreRoundTripAndPersistence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "bd.db")
	s := openTestSQLite(t, path)

	if err := s.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if w, _ := s.ListWishes(ctx, 10); w == nil || len(w) != 0 {
		t.Fatalf("empty wishes = %#v, want non-nil empty slice", w)
	}
	if sc, _ := s.TopScores(ctx, 10); sc == nil || len(sc) != 0 {
		t.Fatalf("empty scores = %#v, want non-nil empty slice", sc)
	}

	base := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	for i, name := range []string{"kho", "Lưu Thành Danh"} {
		if err := s.AddWish(ctx, Wish{ID: name, Name: name, Message: "HPBD 🎂", CreatedAt: base.Add(time.Duration(i) * time.Minute)}); err != nil {
			t.Fatalf("add wish: %v", err)
		}
	}
	scores := []Score{
		{ID: "a", Name: "a", Score: 50, CreatedAt: base},
		{ID: "b", Name: "b", Score: 90, CreatedAt: base.Add(time.Minute)},
		{ID: "c", Name: "c", Score: 50, CreatedAt: base.Add(2 * time.Minute)},
	}
	for _, sc := range scores {
		if err := s.AddScore(ctx, sc); err != nil {
			t.Fatalf("add score: %v", err)
		}
	}
	_ = s.Close()

	// Reopen: data must survive a process restart (the whole point vs the memory store).
	s = openTestSQLite(t, path)
	defer s.Close()

	wishes, err := s.ListWishes(ctx, 10)
	if err != nil || len(wishes) != 2 || wishes[0].Name != "Lưu Thành Danh" || !wishes[0].CreatedAt.Equal(base.Add(time.Minute)) {
		t.Fatalf("wishes after reopen = %+v, err %v; want newest first, Vietnamese intact", wishes, err)
	}
	top, err := s.TopScores(ctx, 2)
	if err != nil || len(top) != 2 || top[0].ID != "b" || top[1].ID != "a" {
		t.Fatalf("top scores = %+v, err %v; want b then a (tie broken by oldest)", top, err)
	}
}

func TestNewStorePicksSQLiteWhenPathSet(t *testing.T) {
	cfg := Config{SQLitePath: filepath.Join(t.TempDir(), "bd.db")}
	store, closeStore, err := newStore(context.Background(), cfg)
	if err != nil {
		t.Fatalf("newStore: %v", err)
	}
	defer closeStore()
	if _, ok := store.(*sqlStore); !ok {
		t.Fatalf("store = %T, want *sqlStore", store)
	}
}
