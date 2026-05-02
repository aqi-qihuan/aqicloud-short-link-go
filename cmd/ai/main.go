package main

import (
	"log"
	"os"

	aiconfig "github.com/aqi/aqicloud-short-link-go/internal/ai/config"
	"github.com/aqi/aqicloud-short-link-go/internal/ai/handler"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	port := getEnv("PORT", "8006")

	cfg := aiconfig.DefaultConfig()
	aiHandler := handler.NewAIHandler(cfg)

	r := gin.Default()
	r.Use(middleware.CorsMiddleware())

	api := r.Group("/api")
	{
		ai := api.Group("/ai/v1")
		ai.Use(interceptor.LoginInterceptor())
		{
			ai.POST("/recommend", aiHandler.Recommend)
			ai.POST("/analytics", aiHandler.Analytics)
			ai.POST("/check_safety", aiHandler.CheckSafety)
		}
	}

	log.Printf("AI service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("AI service start failed: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
