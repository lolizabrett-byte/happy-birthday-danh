package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Wish is a guestbook entry.
type Wish struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// Score is a leaderboard entry for the egg mini-game.
type Score struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is the persistence boundary; memoryStore (dev/tests) and tursoStore (prod) implement it.
type Store interface {
	ListWishes(ctx context.Context, limit int) ([]Wish, error)
	AddWish(ctx context.Context, w Wish) error
	TopScores(ctx context.Context, limit int) ([]Score, error)
	AddScore(ctx context.Context, s Score) error
	Ping(ctx context.Context) error
}

// newID returns 16 random bytes hex-encoded.
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// nowUTC returns the current time truncated to milliseconds (matches DB storage precision).
func nowUTC() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }
