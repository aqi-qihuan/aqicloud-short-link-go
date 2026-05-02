package main

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/aqi/aqicloud-short-link-go/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	port := getEnv("PORT", "8888")
	linkAddr := getEnv("LINK_SERVICE", "http://localhost:8003")
	dataAddr := getEnv("DATA_SERVICE", "http://localhost:8002")
	accountAddr := getEnv("ACCOUNT_SERVICE", "http://localhost:8001")
	shopAddr := getEnv("SHOP_SERVICE", "http://localhost:8005")
	aiAddr := getEnv("AI_SERVICE", "http://localhost:8006")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.RateLimiter(1000, 2000)) // 1000 req/s per IP, burst 2000

	// Route 1: /* -> link-service (short link redirect, highest priority)
	r.Any("/:shortLinkCode", reverseProxy(linkAddr))

	// Route 2-5: /xxx-server/** -> service (strip prefix)
	r.Any("/link-server/*path", stripPrefixProxy("/link-server", linkAddr))
	r.Any("/data-server/*path", stripPrefixProxy("/data-server", dataAddr))
	r.Any("/account-server/*path", stripPrefixProxy("/account-server", accountAddr))
	r.Any("/shop-server/*path", stripPrefixProxy("/shop-server", shopAddr))
	r.Any("/ai-server/*path", stripPrefixProxy("/ai-server", aiAddr))

	log.Printf("Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("gateway start failed: %v", err)
	}
}

func reverseProxy(target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid upstream target: %s, error: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func stripPrefixProxy(prefix string, target string) gin.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid upstream target: %s, error: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	return func(c *gin.Context) {
		originalPath := c.Request.URL.Path
		c.Request.URL.Path = strings.TrimPrefix(originalPath, prefix)
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Request.URL.Path = originalPath // 恢复原始路径
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
