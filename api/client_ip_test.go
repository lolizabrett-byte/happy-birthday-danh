package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPExtractor(t *testing.T) {
	// Render-shaped chain: client, Cloudflare edge (172.71/16 is public), Render LB (private).
	const renderXFF = "203.0.113.5, 172.71.1.2, 10.1.1.1"
	tests := []struct {
		name     string
		onRender bool
		hdr      map[string]string
		want     string
	}{
		{"render prefers True-Client-IP", true, map[string]string{"X-Forwarded-For": renderXFF, "True-Client-IP": "203.0.113.5", "CF-Connecting-IP": "198.51.100.9"}, "203.0.113.5"},
		{"render falls back to CF-Connecting-IP", true, map[string]string{"X-Forwarded-For": renderXFF, "CF-Connecting-IP": "203.0.113.5"}, "203.0.113.5"},
		{"render skips invalid True-Client-IP", true, map[string]string{"True-Client-IP": "garbage", "CF-Connecting-IP": "203.0.113.5"}, "203.0.113.5"},
		{"render without CF headers uses XFF extractor", true, map[string]string{"X-Forwarded-For": "203.0.113.5"}, "203.0.113.5"},
		{"render ipv6 header", true, map[string]string{"True-Client-IP": "2001:db8::1"}, "2001:db8::1"},
		{"off render ignores spoofable CF headers", false, map[string]string{"X-Forwarded-For": "203.0.113.5", "True-Client-IP": "6.6.6.6", "CF-Connecting-IP": "7.7.7.7"}, "203.0.113.5"},
		{"off render XFF through default extractor", false, map[string]string{"X-Forwarded-For": renderXFF}, "172.71.1.2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "10.0.0.1:4321"
			for k, v := range tc.hdr {
				req.Header.Set(k, v)
			}
			if got := clientIPExtractor(tc.onRender)(req); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
