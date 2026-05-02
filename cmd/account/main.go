package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aqi/aqicloud-short-link-go/internal/account/config"
	"github.com/aqi/aqicloud-short-link-go/internal/account/controller"
	"github.com/aqi/aqicloud-short-link-go/internal/account/listener"
	"github.com/aqi/aqicloud-short-link-go/internal/account/service"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/middleware"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/common/sms"
	"github.com/aqi/aqicloud-short-link-go/internal/common/storage"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	port := getEnv("PORT", "8001")
	mysqlHost := getEnv("MYSQL_HOST", "127.0.0.1")
	mysqlPort := getEnv("MYSQL_PORT", "3306")
	mysqlUser := getEnv("MYSQL_USER", "root")
	mysqlPwd := getEnv("MYSQL_PWD", "root")
	redisHost := getEnv("REDIS_HOST", "127.0.0.1")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPwd := getEnv("REDIS_PWD", "")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	storageType := getEnv("STORAGE_TYPE", "local") // "local" or "minio"

	// Account DB
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/aqicloud_account?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlUser, mysqlPwd, mysqlHost, mysqlPort)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect account db failed: %v", err)
	}

	// Traffic tables (traffic_0, traffic_1) are in the same aqicloud_account database
	// Sharding is done at the application layer by table name
	trafficDBs := []*gorm.DB{db, db}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPwd,
	})

	// RabbitMQ
	rmq, err := mq.NewRabbitMQ(rabbitURL)
	if err != nil {
		log.Printf("rabbitmq connect failed: %v (MQ features disabled)", err)
	}

	// Setup MQ exchanges and queues
	if rmq != nil {
		config.SetupExchangesAndQueues(rmq)
	}

	// SMS provider
	smsProvider := getEnv("SMS_PROVIDER", "log") // "log", "alibaba", "tencent"
	smsProv := sms.NewProvider(smsProvider, map[string]string{
		"access_key":    getEnv("SMS_ACCESS_KEY", ""),
		"access_secret": getEnv("SMS_ACCESS_SECRET", ""),
		"sign_name":     getEnv("SMS_SIGN_NAME", "AqiCloud"),
		"app_id":        getEnv("SMS_APP_ID", ""),
		"secret_id":     getEnv("SMS_SECRET_ID", ""),
		"secret_key":    getEnv("SMS_SECRET_KEY", ""),
	})

	// Services
	notifySvc := service.NewNotifyService(rdb, smsProv)
	accountSvc := service.NewAccountService(db, rmq, notifySvc)
	trafficSvc := service.NewTrafficService(trafficDBs, rdb, rmq)

	// Start MQ listeners
	if rmq != nil {
		listener.StartTrafficListeners(rmq, trafficSvc)
		log.Println("Traffic MQ listeners started")
	}

	// File storage
	var store storage.Storage
	switch storageType {
	case "minio":
		minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
		minioBucket := getEnv("MINIO_BUCKET", "aqicloud")
		minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
		minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
		minioPublicURL := getEnv("MINIO_PUBLIC_URL", "http://localhost:9000/aqicloud")
		useSSL := getEnv("MINIO_USE_SSL", "false") == "true"
		store = storage.NewMinIOStorage(minioEndpoint, minioBucket, minioAccessKey, minioSecretKey, useSSL, minioPublicURL)
		log.Printf("Using MinIO storage: %s/%s", minioEndpoint, minioBucket)
	default:
		localPath := getEnv("UPLOAD_PATH", "/data/uploads")
		localURL := getEnv("UPLOAD_URL", fmt.Sprintf("http://localhost:%s/uploads", port))
		store = storage.NewLocalStorage(localPath, localURL)
		log.Printf("Using local storage: %s", localPath)
	}

	// Controllers
	accountCtrl := controller.NewAccountController(accountSvc, store)
	trafficCtrl := controller.NewTrafficController(trafficSvc)

	// Gin router
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())

	// Serve uploaded files (local storage only)
	if storageType == "local" {
		localPath := getEnv("UPLOAD_PATH", "/data/uploads")
		r.Static("/uploads", localPath)
	}

	// Public endpoints (no login required)
	r.POST("/api/account/v1/register", accountCtrl.Register)
	r.POST("/api/account/v1/login", accountCtrl.Login)
	r.POST("/api/account/v1/upload", accountCtrl.Upload)
	r.GET("/api/account/v1/captcha", accountCtrl.Captcha(notifySvc))
	r.POST("/api/account/v1/send_code", accountCtrl.SendCode(notifySvc))
	r.POST("/api/traffic/v1/reduce", trafficCtrl.Reduce) // RPC internal, protected by rpc-token middleware

	// Protected endpoints (login required)
	auth := r.Group("")
	auth.Use(interceptor.LoginInterceptor())
	{
		auth.GET("/api/account/v1/detail", accountCtrl.Detail)
		auth.GET("/api/traffic/v1/page", trafficCtrl.Page)
		auth.GET("/api/traffic/v1/detail/:trafficId", trafficCtrl.Detail)
	}

	log.Printf("Account service starting on port %s", port)
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
