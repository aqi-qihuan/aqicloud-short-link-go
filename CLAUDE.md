# AqiCloud Short-Link Go

Go rewrite of the Java short-link microservice platform. 5 services, 32+ API endpoints, full middleware stack.

## Build

```bash
# Build all services
go build ./cmd/...

# Build individual services
go build ./cmd/gateway/    # :8888
go build ./cmd/account/    # :8001
go build ./cmd/data/       # :8002
go build ./cmd/link/       # :8003
go build ./cmd/shop/       # :8005
go build ./cmd/ai/         # :8006

# Run
go run ./cmd/link/main.go

# Docker (all services + middleware)
docker-compose up -d
```

## Architecture

```
cmd/
  gateway/main.go       # Reverse proxy, CORS, rate limiting
  account/main.go       # User auth, traffic management, SMS
  data/main.go          # ClickHouse visit statistics queries
  link/main.go          # Short link CRUD, redirect (hot path)
  shop/main.go          # Products, orders, payment callbacks
  ai/main.go            # AI features (recommendation, analytics, URL safety)

internal/
  common/               # Shared code across all services
    config/             # MySQL, Redis connection factories
    constant/           # Redis key patterns
    enums/              # 30+ business error codes, state enums
    interceptor/        # JWT login middleware (gin.Context-based)
    middleware/          # CORS, rate limiter, RPC token
    model/              # EventMessage, LoginUser, LogRecord
    mq/                 # RabbitMQ (amqp091-go) + Kafka (kafka-go) wrappers
    response/           # JsonData{Code, Data, Msg} response format
    util/               # JWT, MurmurHash3, Base62, MD5, snowflake ID
  link/
    component/          # Short link code generation (MurmurHash3 + Base62)
    config/             # RabbitMQ exchange/queue setup
    controller/         # ShortLink, LinkGroup, Domain, LinkApi controllers
    listener/           # 7 MQ consumers (add/del/update x link/mapping + error)
    model/              # ShortLinkDO, GroupCodeMappingDO, LinkGroupDO, DomainDO
    request/            # Request DTOs
    service/            # ShortLinkService (MQ handler with collision retry)
    sharding/           # Application-layer DB/table routing
    vo/                 # View objects
  account/              # Account, Traffic, Notify controllers + services
  shop/                 # Product, Order, Callback controllers + services
  data/                 # ClickHouse visit stats service
  ai/                   # (Phase 6) AI features with Eino + trpc-agent-go
```

## Key Compatibility Points

- **MurmurHash3**: Guava-compatible UTF-16 LE encoding (`internal/common/util/hash.go`)
- **Java String.hashCode()**: Used for sharding routing (`internal/common/util/hash.go`)
- **JWT**: HS256, secret via `JWT_SECRET` env var, prefix=`dcloud-link`, 7-day expiry
- **MD5-crypt passwords**: `$1$` + 8-char salt, uses `GehirnInc/crypt`
- **Base62 charset**: `0-9a-zA-Z` (variable-length, not zero-padded)
- **MD5 output**: Uppercase 32-char hex

## Sharding Strategy

- **short_link**: DB by code[0], table by code[last] (3 DBs: 0/1/a, 2 tables: 0/a)
- **group_code_mapping**: DB by account_no%2, table by group_id%2
- **traffic**: DB by account_no%2 (tables: traffic_0, traffic_1)

## RabbitMQ Topology

- `short_link.event.exchange` - 6 queues (add/del/update x link/mapping) + error
- `traffic.event.exchange` - 3 queues (free_init, release delay/dead-letter, order_traffic) + error
- `order.event.exchange` - 4 queues (close delay/dead-letter, update, traffic) + error

## Environment Variables

| Variable | Default | Service |
|----------|---------|---------|
| JWT_SECRET | (required) | All (JWT signing) |
| PORT | 8001/8002/8003/8005/8006/8888 | All |
| MYSQL_HOST | 127.0.0.1 | account, link, shop |
| MYSQL_PORT | 3306 | account, link, shop |
| MYSQL_USER | root | account, link, shop |
| MYSQL_PWD | root | account, link, shop |
| REDIS_HOST | 127.0.0.1 | account, link, shop |
| REDIS_PORT | 6379 | account, link, shop |
| RABBITMQ_URL | amqp://guest:guest@localhost:5672/ | account, link, shop |
| KAFKA_BROKERS | localhost:9092 | link |
| CLICKHOUSE_HOST | 127.0.0.1 | data |
| CLICKHOUSE_PORT | 9000 | data |
| ACCOUNT_SERVICE | http://localhost:8001 | link |
| RPC_TOKEN | rpc-token-default | link |
| STORAGE_TYPE | local | account |
| MINIO_ENDPOINT | minio:9000 | account |
| MINIO_BUCKET | aqicloud | account |
| MINIO_ACCESS_KEY | minioadmin | account |
| MINIO_SECRET_KEY | minioadmin | account |
| MINIO_PUBLIC_URL | http://localhost:9000/aqicloud | account |
| SMS_PROVIDER | log | account |
| ALERT_TYPE | log | link |
| ALERT_WEBHOOK_URL | — | link |
| AI_PROVIDER | doubao | ai |
| AI_API_KEY | — | ai |
| AI_BASE_URL | — | ai |
| AI_MODEL | doubao-pro-32k | ai |

## Databases

- `aqicloud_account` - account table
- `aqicloud_traffic_0`, `aqicloud_traffic_1` - traffic tables (or in aqicloud_account)
- `aqicloud_link_0`, `aqicloud_link_1`, `aqicloud_link_a` - short_link, group_code_mapping, link_group
- `aqicloud_shop` - product, product_order
- ClickHouse `visit_stats` - Flink-aggregated visit data

## Dependencies

- Go 1.22+, Gin, GORM, go-redis/v9, amqp091-go, kafka-go, golang-jwt/v5, sonyflake, murmur3, GehirnInc/crypt
- ClickHouse driver (`github.com/ClickHouse/clickhouse-go/v2`) - needs `go get` when network available
