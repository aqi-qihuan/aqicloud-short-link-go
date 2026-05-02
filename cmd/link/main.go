package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aqi/aqicloud-short-link-go/internal/common/alert"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/middleware"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/link/config"
	"github.com/aqi/aqicloud-short-link-go/internal/link/controller"
	"github.com/aqi/aqicloud-short-link-go/internal/link/listener"
	"github.com/aqi/aqicloud-short-link-go/internal/link/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Config from env (will be replaced by Nacos later)
	port := getEnv("PORT", "8003")
	mysqlHost := getEnv("MYSQL_HOST", "127.0.0.1")
	mysqlPort := getEnv("MYSQL_PORT", "3306")
	mysqlUser := getEnv("MYSQL_USER", "root")
	mysqlPwd := getEnv("MYSQL_PWD", "root")
	redisHost := getEnv("REDIS_HOST", "127.0.0.1")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPwd := getEnv("REDIS_PWD", "")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	accountAddr := getEnv("ACCOUNT_SERVICE", "http://localhost:8001")
	rpcToken := getEnv("RPC_TOKEN", "rpc-token-default")

	// Connect to 3 MySQL datasources (sharded)
	dsn0 := fmt.Sprintf("%s:%s@tcp(%s:%s)/aqicloud_link_0?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlPort)
	dsn1 := fmt.Sprintf("%s:%s@tcp(%s:%s)/aqicloud_link_1?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlPort)
	dsnA := fmt.Sprintf("%s:%s@tcp(%s:%s)/aqicloud_link_a?charset=utf8mb4&parseTime=True&loc=Local", mysqlUser, mysqlPwd, mysqlHost, mysqlPort)

	db0, err := gorm.Open(mysql.Open(dsn0), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect link_0 failed: %v", err)
	}
	db1, err := gorm.Open(mysql.Open(dsn1), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect link_1 failed: %v", err)
	}
	dbA, err := gorm.Open(mysql.Open(dsnA), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect link_a failed: %v", err)
	}
	dbs := []*gorm.DB{db0, db1, dbA}

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

	// Kafka producer
	kafka := mq.NewKafkaProducer([]string{kafkaBrokers}, "ods_link_visit_topic")

	// Setup MQ exchanges and queues
	if rmq != nil {
		config.SetupExchangesAndQueues(rmq)
	}

	// Service layer (used by MQ listeners)
	shortLinkSvc := service.NewShortLinkService(dbs, rdb, accountAddr, rpcToken)

	// Alert notification
	alerter := alert.NewAlerter()

	// Start MQ listeners
	if rmq != nil {
		listener.StartAddLinkListener(rmq, shortLinkSvc)
		listener.StartAddMappingListener(rmq, shortLinkSvc)
		listener.StartDelLinkListener(rmq, shortLinkSvc)
		listener.StartDelMappingListener(rmq, shortLinkSvc)
		listener.StartUpdateLinkListener(rmq, shortLinkSvc)
		listener.StartUpdateMappingListener(rmq, shortLinkSvc)
		listener.StartErrorListener(rmq, alerter)
		log.Println("All MQ listeners started")
	}

	// Controllers
	shortLinkCtrl := controller.NewShortLinkController(dbs, rdb, rmq, kafka)
	linkGroupCtrl := controller.NewLinkGroupController(dbs[:2]) // ds0, ds1 only
	domainCtrl := controller.NewDomainController(db0)            // domain table only in ds0
	linkApiCtrl := controller.NewLinkApiController(dbs, kafka)

	// Gin router
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())

	// Short link redirect (hot path, no auth)
	r.GET("/:shortLinkCode", linkApiCtrl.Dispatch)

	// API routes
	api := r.Group("/api")
	{
		// RPC internal (no login required)
		api.GET("/link/v1/check", shortLinkCtrl.Check)

		// Protected routes (login required)
		auth := api.Group("")
		auth.Use(interceptor.LoginInterceptor())
		{
			link := auth.Group("/link/v1")
			{
				link.POST("/add", shortLinkCtrl.Add)
				link.POST("/page", shortLinkCtrl.Page)
				link.POST("/detail", shortLinkCtrl.Detail)
				link.POST("/del", shortLinkCtrl.Del)
				link.POST("/update", shortLinkCtrl.Update)
			}
			group := auth.Group("/group/v1")
			{
				group.POST("/add", linkGroupCtrl.Add)
				group.DELETE("/del/:group_id", linkGroupCtrl.Del)
				group.GET("/detail/:group_id", linkGroupCtrl.Detail)
				group.GET("/list", linkGroupCtrl.List)
				group.PUT("/update", linkGroupCtrl.Update)
			}
			domain := auth.Group("/domain/v1")
			{
				domain.GET("/list", domainCtrl.List)
			}
		}
	}

	log.Printf("Link service starting on port %s", port)
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
