package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

const testOrigin = "http://localhost:8000"

func testConfig() Config {
	return Config{
		AllowedOrigins: []string{testOrigin},
		RateLimit:      rate.Limit(1000),
		RateBurst:      1000,
	}
}

func newTestServer(t *testing.T) (*echo.Echo, *memoryStore) {
	t.Helper()
	ms := newMemoryStore()
	return newServer(ms, testConfig()), ms
}

// do sends a request through the router; ip is placed in X-Forwarded-For behind a private proxy address.
func do(e *echo.Echo, method, path, body, ip string, hdr ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "10.0.0.1:4321"
	req.Header.Set("X-Forwarded-For", ip)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("bad json %q: %v", rec.Body.String(), err)
	}
	return v
}

type wishList struct {
	Wishes []Wish `json:"wishes"`
}
type scoreList struct {
	Scores []Score `json:"scores"`
}

func TestPostWishValid(t *testing.T) {
	e, _ := newTestServer(t)
	rec := do(e, http.MethodPost, "/api/wishes", `{"name":" kho ","message":"HPBD Danh","website":""}`, "1.1.1.1")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body)
	}
	w := decode[Wish](t, rec)
	if len(w.ID) != 32 || w.Name != "kho" || w.Message != "HPBD Danh" || w.CreatedAt.IsZero() {
		t.Fatalf("bad wish %+v", w)
	}
	if !strings.Contains(rec.Body.String(), `"createdAt"`) {
		t.Fatalf("missing createdAt key: %s", rec.Body)
	}
	list := decode[wishList](t, do(e, http.MethodGet, "/api/wishes", "", "1.1.1.1"))
	if len(list.Wishes) != 1 || list.Wishes[0].ID != w.ID {
		t.Fatalf("list %+v", list)
	}
}

func TestPostWishInvalid(t *testing.T) {
	e, _ := newTestServer(t)
	for name, body := range map[string]string{
		"bad json":      `{not json`,
		"empty name":    `{"name":"  ","message":"hi"}`,
		"empty message": `{"name":"kho","message":""}`,
		"long message":  `{"name":"kho","message":"` + strings.Repeat("a", 281) + `"}`,
	} {
		rec := do(e, http.MethodPost, "/api/wishes", body, "1.1.1.1")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", name, rec.Code)
		}
		if decode[map[string]string](t, rec)["error"] == "" {
			t.Errorf("%s: missing error field: %s", name, rec.Body)
		}
	}
	if rec := do(e, http.MethodPost, "/api/wishes", `{"name":"a","message":"b"}`, "1.1.1.1", "Content-Type", "text/plain"); rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("non-json content type: status %d", rec.Code)
	}
}

func TestHoneypotNotStored(t *testing.T) {
	e, _ := newTestServer(t)
	rec := do(e, http.MethodPost, "/api/wishes", `{"name":"bot","message":"spam","website":"http://x"}`, "1.1.1.1")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d", rec.Code)
	}
	if w := decode[Wish](t, rec); w.ID == "" {
		t.Fatalf("fake wish has no id: %+v", w)
	}
	if list := decode[wishList](t, do(e, http.MethodGet, "/api/wishes", "", "1.1.1.1")); len(list.Wishes) != 0 {
		t.Fatalf("honeypot stored: %+v", list)
	}
}

func TestWishLimitClamp(t *testing.T) {
	e, ms := newTestServer(t)
	for i := 0; i < 120; i++ {
		_ = ms.AddWish(context.Background(), Wish{ID: fmt.Sprint(i), Name: "n", Message: "m", CreatedAt: nowUTC()})
	}
	for q, want := range map[string]int{"": 50, "?limit=0": 50, "?limit=-3": 50, "?limit=abc": 50, "?limit=7": 7, "?limit=500": 100} {
		list := decode[wishList](t, do(e, http.MethodGet, "/api/wishes"+q, "", "1.1.1.1"))
		if len(list.Wishes) != want {
			t.Errorf("limit %q: got %d want %d", q, len(list.Wishes), want)
		}
	}
	if list := decode[wishList](t, do(e, http.MethodGet, "/api/wishes?limit=3", "", "1.1.1.1")); list.Wishes[0].ID != "119" {
		t.Errorf("not newest first: %+v", list.Wishes)
	}
}

func TestScoresSortedTiesOldestFirst(t *testing.T) {
	e, _ := newTestServer(t)
	for _, s := range []struct {
		name  string
		score int
	}{{"a", 10}, {"b", 30}, {"c", 10}, {"d", 20}} {
		body := fmt.Sprintf(`{"name":%q,"score":%d}`, s.name, s.score)
		if rec := do(e, http.MethodPost, "/api/scores", body, "1.1.1.1"); rec.Code != http.StatusCreated {
			t.Fatalf("post %s: %d %s", s.name, rec.Code, rec.Body)
		}
	}
	got := ""
	for _, s := range decode[scoreList](t, do(e, http.MethodGet, "/api/scores", "", "1.1.1.1")).Scores {
		got += s.Name
	}
	if got != "bdac" {
		t.Fatalf("order %q, want bdac", got)
	}
	if n := len(decode[scoreList](t, do(e, http.MethodGet, "/api/scores?limit=2", "", "1.1.1.1")).Scores); n != 2 {
		t.Fatalf("limit=2 returned %d", n)
	}
}

func TestPostScoreInvalid(t *testing.T) {
	e, _ := newTestServer(t)
	for name, body := range map[string]string{
		"negative":  `{"name":"a","score":-1}`,
		"too big":   `{"name":"a","score":1001}`,
		"long name": `{"name":"` + strings.Repeat("a", 21) + `","score":5}`,
		"str score": `{"name":"a","score":"9"}`,
	} {
		if rec := do(e, http.MethodPost, "/api/scores", body, "1.1.1.1"); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", name, rec.Code)
		}
	}
}

func TestBodyOver4KB(t *testing.T) {
	e, _ := newTestServer(t)
	body := `{"name":"kho","message":"` + strings.Repeat("a", 5000) + `"}`
	if rec := do(e, http.MethodPost, "/api/wishes", body, "1.1.1.1"); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRateLimitPostOnly(t *testing.T) {
	cfg := testConfig()
	cfg.RateLimit, cfg.RateBurst = rate.Limit(0.0001), 3 // effectively no refill during the test
	e := newServer(newMemoryStore(), cfg)
	post := func(ip string) int {
		return do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi"}`, ip).Code
	}
	for i := 0; i < 3; i++ {
		if c := post("2.2.2.2"); c != http.StatusCreated {
			t.Fatalf("post %d: %d", i+1, c)
		}
	}
	rec := do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi"}`, "2.2.2.2")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("4th post: %d", rec.Code)
	}
	if decode[map[string]string](t, rec)["error"] == "" {
		t.Errorf("429 lacks error field: %s", rec.Body)
	}
	if c := post("3.3.3.3"); c != http.StatusCreated {
		t.Errorf("other ip limited: %d", c)
	}
	for i := 0; i < 10; i++ {
		if c := do(e, http.MethodGet, "/api/wishes", "", "2.2.2.2").Code; c != http.StatusOK {
			t.Fatalf("GET limited: %d", c)
		}
	}
}

func TestCORS(t *testing.T) {
	e, _ := newTestServer(t)
	pre := func(origin string) *httptest.ResponseRecorder {
		return do(e, http.MethodOptions, "/api/wishes", "", "1.1.1.1",
			"Origin", origin, "Access-Control-Request-Method", "POST", "Access-Control-Request-Headers", "Content-Type")
	}
	if rec := pre(testOrigin); rec.Header().Get("Access-Control-Allow-Origin") != testOrigin {
		t.Errorf("allowed origin: ACAO=%q code=%d", rec.Header().Get("Access-Control-Allow-Origin"), rec.Code)
	}
	if rec := pre("https://evil.example"); rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("disallowed origin got ACAO=%q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

type failStore struct{ *memoryStore }

func (failStore) Ping(context.Context) error          { return errors.New("down") }
func (failStore) AddWish(context.Context, Wish) error { return errors.New("down") }

func TestStoreFailures(t *testing.T) {
	e := newServer(failStore{newMemoryStore()}, testConfig())
	if rec := do(e, http.MethodGet, "/healthz", "", "1.1.1.1"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("healthz down: %d", rec.Code)
	}
	rec := do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi"}`, "1.1.1.1")
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "down") {
		t.Errorf("store error leaked or wrong status: %d %s", rec.Code, rec.Body)
	}
}

func TestHealthz(t *testing.T) {
	e, _ := newTestServer(t)
	rec := do(e, http.MethodGet, "/healthz", "", "1.1.1.1")
	if rec.Code != http.StatusOK || !decode[map[string]bool](t, rec)["ok"] {
		t.Fatalf("healthz: %d %s", rec.Code, rec.Body)
	}
}

func TestConcurrentAddWish(t *testing.T) {
	ms := newMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = ms.AddWish(context.Background(), Wish{ID: fmt.Sprint(i), Name: "n", Message: "m", CreatedAt: time.Now().UTC()})
		}(i)
	}
	wg.Wait()
	got, err := ms.ListWishes(context.Background(), 100)
	if err != nil || len(got) != 50 {
		t.Fatalf("got %d rows, err %v", len(got), err)
	}
}

func TestSeparateRateLimitBuckets(t *testing.T) {
	cfg := testConfig()
	cfg.RateLimit, cfg.RateBurst = rate.Limit(0.0001), 3
	e := newServer(newMemoryStore(), cfg)
	for i := 0; i < 3; i++ {
		do(e, http.MethodPost, "/api/scores", `{"name":"kho","score":5}`, "4.4.4.4")
	}
	if c := do(e, http.MethodPost, "/api/scores", `{"name":"kho","score":5}`, "4.4.4.4").Code; c != http.StatusTooManyRequests {
		t.Fatalf("score bucket not exhausted: %d", c)
	}
	if c := do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi"}`, "4.4.4.4").Code; c != http.StatusCreated {
		t.Fatalf("wish blocked by score bucket: %d", c)
	}
}

func TestRateLimitKeyOnRender(t *testing.T) {
	cfg := testConfig()
	cfg.RateLimit, cfg.RateBurst, cfg.OnRender = rate.Limit(0.0001), 1, true
	e := newServer(newMemoryStore(), cfg)
	const xff = "203.0.113.5, 172.71.1.2, 10.1.1.1" // same Cloudflare edge for everyone
	post := func(client string) int {
		return do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi"}`, xff, "True-Client-IP", client).Code
	}
	if c := post("198.51.100.1"); c != http.StatusCreated {
		t.Fatalf("first: %d", c)
	}
	if c := post("198.51.100.1"); c != http.StatusTooManyRequests {
		t.Fatalf("same client not limited: %d", c)
	}
	if c := post("198.51.100.2"); c != http.StatusCreated {
		t.Fatalf("different client shares bucket: %d", c)
	}
}

func TestRequestLogIncludesRemoteIP(t *testing.T) {
	var buf strings.Builder
	cfg := testConfig()
	cfg.Logger = slog.New(slog.NewTextHandler(&buf, nil))
	e := newServer(newMemoryStore(), cfg)
	do(e, http.MethodGet, "/healthz", "", "203.0.113.7")
	if !strings.Contains(buf.String(), "remote_ip=203.0.113.7") {
		t.Fatalf("log lacks remote_ip: %q", buf.String())
	}
}

func TestEmptyListsAreArrays(t *testing.T) {
	e, _ := newTestServer(t)
	if b := do(e, http.MethodGet, "/api/scores", "", "1.1.1.1").Body.String(); strings.TrimSpace(b) != `{"scores":[]}` {
		t.Errorf("scores: %s", b)
	}
	if b := do(e, http.MethodGet, "/api/wishes", "", "1.1.1.1").Body.String(); strings.TrimSpace(b) != `{"wishes":[]}` {
		t.Errorf("wishes: %s", b)
	}
}

func TestCORSHeadersOn413AndMaxAge(t *testing.T) {
	e, _ := newTestServer(t)
	body := `{"name":"kho","message":"` + strings.Repeat("a", 5000) + `"}`
	rec := do(e, http.MethodPost, "/api/wishes", body, "1.1.1.1", "Origin", testOrigin)
	if rec.Code != http.StatusRequestEntityTooLarge || rec.Header().Get("Access-Control-Allow-Origin") != testOrigin {
		t.Errorf("413 without ACAO: %d %v", rec.Code, rec.Header())
	}
	pre := do(e, http.MethodOptions, "/api/wishes", "", "1.1.1.1", "Origin", testOrigin, "Access-Control-Request-Method", "POST")
	if got := pre.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Errorf("max-age = %q", got)
	}
}

func TestPostScoreMissingScore(t *testing.T) {
	e, _ := newTestServer(t)
	if rec := do(e, http.MethodPost, "/api/scores", `{"name":"kho"}`, "1.1.1.1"); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing score: %d %s", rec.Code, rec.Body)
	}
	if rec := do(e, http.MethodPost, "/api/scores", `{"name":"kho","score":0}`, "1.1.1.1"); rec.Code != http.StatusCreated {
		t.Fatalf("explicit zero rejected: %d", rec.Code)
	}
}

func TestHoneypotValidatesFirstAndEchoesCleanValues(t *testing.T) {
	e, _ := newTestServer(t)
	if rec := do(e, http.MethodPost, "/api/wishes", `{"name":"","message":"x","website":"http://x"}`, "1.1.1.1"); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid honeypot request: %d", rec.Code)
	}
	rec := do(e, http.MethodPost, "/api/wishes", `{"name":"  bot ","message":" spam ","website":"http://x"}`, "1.1.1.1")
	if w := decode[Wish](t, rec); rec.Code != http.StatusCreated || w.Name != "bot" || w.Message != "spam" {
		t.Fatalf("honeypot response %d %+v", rec.Code, w)
	}
}

func TestWhitespaceWebsiteIsNotBot(t *testing.T) {
	e, _ := newTestServer(t)
	do(e, http.MethodPost, "/api/wishes", `{"name":"kho","message":"hi","website":"  \t"}`, "1.1.1.1")
	if list := decode[wishList](t, do(e, http.MethodGet, "/api/wishes", "", "1.1.1.1")); len(list.Wishes) != 1 {
		t.Fatalf("whitespace website treated as bot: %+v", list)
	}
}

func TestHeadHealthz(t *testing.T) {
	e, _ := newTestServer(t)
	if rec := do(e, http.MethodHead, "/healthz", "", "1.1.1.1"); rec.Code != http.StatusOK {
		t.Fatalf("HEAD /healthz: %d", rec.Code)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", " , ,")
	if _, err := loadConfig(); err == nil {
		t.Error("empty origin list must be an error")
	}
	t.Setenv("ALLOWED_ORIGINS", "https://a.github.io/, https://b.example")
	t.Setenv("RENDER", "true")
	cfg, err := loadConfig()
	if err != nil || len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "https://a.github.io" || !cfg.OnRender {
		t.Fatalf("cfg %+v err %v", cfg, err)
	}
}

func TestNewStoreRequiresTursoOnRender(t *testing.T) {
	if _, _, err := newStore(context.Background(), Config{OnRender: true}); err == nil {
		t.Error("render without TURSO_DATABASE_URL must fail")
	}
	st, closeFn, err := newStore(context.Background(), Config{})
	if err != nil || st == nil {
		t.Fatalf("dev memory store: %v", err)
	}
	closeFn()
}
