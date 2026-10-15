package main

import (
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type handlers struct {
	store Store
}

// newServer wires middleware and routes; it never starts listening (main.go does).
func newServer(store Store, cfg Config) *echo.Echo {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	e := echo.New()
	e.HideBanner, e.HidePort = true, true
	e.IPExtractor = clientIPExtractor(cfg.OnRender)
	e.HTTPErrorHandler = jsonErrorHandler

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod: true, LogURI: true, LogStatus: true, LogLatency: true, LogRemoteIP: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request", "method", v.Method, "path", v.URI, "status", v.Status,
				"latency", v.Latency, "remote_ip", v.RemoteIP)
			return nil
		},
	}))
	// CORS first so error responses (e.g. 413 from BodyLimit) still carry Access-Control-Allow-Origin.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType},
		MaxAge:       600,
	}))
	e.Use(middleware.BodyLimit("4K"))

	h := &handlers{store: store}
	e.GET("/healthz", h.health)
	e.HEAD("/healthz", h.health)
	e.GET("/api/wishes", h.listWishes)
	e.POST("/api/wishes", h.createWish, newLimiter(cfg)) // one bucket per POST route
	e.GET("/api/scores", h.listScores)
	e.POST("/api/scores", h.createScore, newLimiter(cfg))
	return e
}

func newLimiter(cfg Config) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate: cfg.RateLimit, Burst: cfg.RateBurst,
		}),
	})
}

func (h *handlers) health(c echo.Context) error {
	if err := h.store.Ping(c.Request().Context()); err != nil {
		c.Logger().Errorf("healthz ping: %v", err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, "store unavailable")
	}
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

// jsonErrorHandler renders every error as {"error": "..."}; non-HTTP errors become a generic 500.
func jsonErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	code, msg := http.StatusInternalServerError, "internal server error"
	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		if s, ok := he.Message.(string); ok {
			msg = s
		} else {
			msg = http.StatusText(code)
		}
	}
	if c.Request().Method == http.MethodHead {
		_ = c.NoContent(code)
		return
	}
	_ = c.JSON(code, map[string]string{"error": msg})
}

// bindJSON enforces a JSON content type and decodes the (already size-limited) body.
func bindJSON(c echo.Context, dst any) error {
	mt, _, _ := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if mt != echo.MIMEApplicationJSON {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "content type must be application/json")
	}
	if err := c.Bind(dst); err != nil {
		var he *echo.HTTPError
		if errors.As(err, &he) && he.Code == http.StatusRequestEntityTooLarge {
			return he
		}
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	return nil
}

// parseLimit reads ?limit=; invalid or out-of-range (<1) -> def, above max -> max.
func parseLimit(c echo.Context, def, max int) int {
	n, err := strconv.Atoi(c.QueryParam("limit"))
	switch {
	case err != nil || n < 1:
		return def
	case n > max:
		return max
	}
	return n
}

func badRequest(err error) error { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }

var errInternal = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
