package controllers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	sameSite := resolveSameSiteMode(c)
	secure := shouldUseSecureCookies(c, sameSite)

	c.SetSameSite(sameSite)
	c.SetCookie("refresh_token", refreshToken, 30*24*3600, "/", "", secure, true)
}

func resolveSameSiteMode(c *gin.Context) http.SameSite {
	origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
	if origin == "" {
		return http.SameSiteLaxMode
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return http.SameSiteLaxMode
	}

	requestHost := forwardedHost(c)
	if requestHost == "" {
		requestHost = c.Request.Host
	}

	if !sameHostname(originURL.Hostname(), requestHost) {
		return http.SameSiteNoneMode
	}

	return http.SameSiteLaxMode
}

func shouldUseSecureCookies(c *gin.Context, sameSite http.SameSite) bool {
	if sameSite == http.SameSiteNoneMode {
		return true
	}

	if c.Request.TLS != nil {
		return true
	}

	return strings.EqualFold(c.Request.Header.Get("X-Forwarded-Proto"), "https")
}

func forwardedHost(c *gin.Context) string {
	host := strings.TrimSpace(c.Request.Header.Get("X-Forwarded-Host"))
	if host == "" {
		return ""
	}

	if idx := strings.Index(host, ","); idx >= 0 {
		host = host[:idx]
	}

	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")

	return host
}

func sameHostname(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(stripPort(b)))
}

func stripPort(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	if strings.HasPrefix(host, "[") {
		if idx := strings.LastIndex(host, "]"); idx >= 0 {
			return host[1:idx]
		}
	}

	if idx := strings.LastIndex(host, ":"); idx >= 0 && strings.Count(host, ":") == 1 {
		return host[:idx]
	}

	return host
}
