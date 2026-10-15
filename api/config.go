package main

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/time/rate"
)

const defaultOrigins = "http://localhost:8000,http://127.0.0.1:8000"

// Config holds runtime settings. Rate limiter fields are injectable so tests stay deterministic.
type Config struct {
	Host           string // bind address; "127.0.0.1" behind a local reverse proxy, "" = all interfaces
	Port           string
	TursoURL       string
	TursoToken     string
	SQLitePath     string // local SQLite file (self-hosted VPS); used when TursoURL is empty
	AllowedOrigins []string
	OnRender       bool       // RENDER env set: behind Cloudflare + Render LB, client IP comes from CF headers
	RateLimit      rate.Limit // tokens per second for POST routes
	RateBurst      int
	Logger         *slog.Logger // nil = discard
}

func loadConfig() (Config, error) {
	origins := splitOrigins(envOr("ALLOWED_ORIGINS", defaultOrigins))
	if len(origins) == 0 {
		return Config{}, errors.New("ALLOWED_ORIGINS contains no origins")
	}
	return Config{
		Host:           strings.TrimSpace(os.Getenv("HOST")),
		Port:           envOr("PORT", "8080"),
		TursoURL:       os.Getenv("TURSO_DATABASE_URL"),
		TursoToken:     os.Getenv("TURSO_AUTH_TOKEN"),
		SQLitePath:     strings.TrimSpace(os.Getenv("SQLITE_PATH")),
		AllowedOrigins: origins,
		OnRender:       os.Getenv("RENDER") != "",
		RateLimit:      rate.Limit(5.0 / 60.0), // 5 requests per minute per IP
		RateBurst:      3,
		Logger:         slog.Default(),
	}, nil
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func splitOrigins(s string) []string {
	var out []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			out = append(out, o)
		}
	}
	return out
}
