package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// clientIPExtractor returns the client IP used for rate limiting and logs.
//
// On Render the request path is client -> Cloudflare -> Render LB, so X-Forwarded-For ends with the
// Cloudflare edge IP (public range), which the default XFF extractor would pick as "the client".
// There True-Client-IP / CF-Connecting-IP carry the real address. Elsewhere those headers are
// client-controlled, so they are ignored and the XFF extractor is used.
func clientIPExtractor(onRender bool) echo.IPExtractor {
	xff := echo.ExtractIPFromXFFHeader()
	if !onRender {
		return xff
	}
	return func(r *http.Request) string {
		for _, h := range []string{"True-Client-IP", "CF-Connecting-IP"} {
			if ip := net.ParseIP(strings.TrimSpace(r.Header.Get(h))); ip != nil {
				return ip.String()
			}
		}
		return xff(r)
	}
}
