package main

import (
	"context"
	"sort"
	"sync"
)

// memoryStore keeps everything in process memory; used for dev (no Turso URL) and tests.
type memoryStore struct {
	mu     sync.RWMutex
	wishes []Wish  // insertion order
	scores []Score // insertion order
}

func newMemoryStore() *memoryStore { return &memoryStore{} }

func (m *memoryStore) ListWishes(_ context.Context, limit int) ([]Wish, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Wish, 0, min(limit, len(m.wishes)))
	for i := len(m.wishes) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, m.wishes[i])
	}
	return out, nil
}

func (m *memoryStore) AddWish(_ context.Context, w Wish) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wishes = append(m.wishes, w)
	return nil
}

func (m *memoryStore) TopScores(_ context.Context, limit int) ([]Score, error) {
	m.mu.RLock()
	sorted := append([]Score{}, m.scores...) // non-nil: empty list must encode as [] not null
	m.mu.RUnlock()
	// Stable over insertion order: score desc, then oldest first.
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted, nil
}

func (m *memoryStore) AddScore(_ context.Context, s Score) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scores = append(m.scores, s)
	return nil
}

func (m *memoryStore) Ping(context.Context) error { return nil }
