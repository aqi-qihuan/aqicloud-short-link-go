package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/middleware"
	"github.com/aqi/aqicloud-short-link-go/internal/data/controller"
	"github.com/aqi/aqicloud-short-link-go/internal/data/service"
	"github.com/gin-gonic/gin"
)

func main() {
	port := getEnv("PORT", "8002")
	chHost := getEnv("CLICKHOUSE_HOST", "127.0.0.1")
	chPort := getEnv("CLICKHOUSE_PORT", "9000")
	chUser := getEnv("CLICKHOUSE_USER", "default")
	chPwd := getEnv("CLICKHOUSE_PWD", "")
	chDB := getEnv("CLICKHOUSE_DB", "default")

	// ClickHouse connection via database/sql
	// Requires: import _ "github.com/ClickHouse/clickhouse-go/v2" at build time
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s", chUser, chPwd, chHost, chPort, chDB)
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("connect clickhouse failed: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping clickhouse failed: %v", err)
	}

	// Service + Controller
	statsSvc := service.NewVisitStatsService(db)
	statsCtrl := controller.NewVisitStatsController(statsSvc)

	// Gin router
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())

	// Protected endpoints
	api := r.Group("/api")
	api.Use(interceptor.LoginInterceptor())
	{
		stats := api.Group("/visit_stats/v1")
		{
			stats.POST("/page_record", statsCtrl.PageRecord)
			stats.POST("/region_day", statsCtrl.RegionDay)
			stats.POST("/trend", statsCtrl.Trend)
			stats.POST("/frequent_ip", statsCtrl.FrequentIP)
			stats.POST("/frequent_referer", statsCtrl.FrequentReferer)
			stats.POST("/device_info", statsCtrl.DeviceInfo)
		}
	}

	log.Printf("Data service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
