# AqiCloud Short-Link 性能分析与压测报告

> 生成时间: 2026-05-02 14:25
> 分析范围: 全部 6 个微服务 + 中间件配置

---

## 一、架构总览

```
                         ┌─────────────────┐
                         │   Gateway :8888  │
                         │  限流 100r/s/IP  │
                         └────────┬────────┘
          ┌──────┬──────┬─────────┼─────────┬──────┐
          ▼      ▼      ▼         ▼         ▼      ▼
      Account  Data    Link     Shop      AI
      :8001   :8002   :8003    :8005     :8006
                │       │
          ┌─────┘  ┌────┼────┬──────────┐
          ▼        ▼    ▼    ▼          ▼
       ClickHouse MySQL Redis RabbitMQ  Kafka
                 (3分库)      (MQ)    (日志)
```

**热路径**: `GET /:shortLinkCode` → Gateway → Link Service → MySQL查询 → 302重定向

---

## 二、已识别的性能瓶颈 (按严重程度排序)

### P0 - 致命级 (直接决定高峰期能否撑住)

#### 1. 热路径无 Redis 缓存 — 每次访问打 MySQL

**位置**: `internal/link/controller/link_api.go:50-56`

```go
// 当前: 每次短链访问都查 MySQL
err := ctrl.dbs[dbIdx].Table(tableName).
    Where("code = ? AND del = 0", code).
    First(&shortLink).Error
```

**问题**: 短链重定向是 QPS 最高的接口。假设高峰期 10,000 QPS，每次都查 MySQL（即使有索引），3 个分库 × 2 张表 = 6 个物理表，单表承受 ~1,667 QPS 的 `SELECT`。MySQL InnoDB 在这个量级下 buffer pool 压力极大，P99 延迟会飙升到 50-100ms。

**影响**: 这是 **性能差的根本原因**。短链服务没有使用已注入的 Redis 客户端做任何缓存。

**优化方案**:
```
Redis 缓存短链映射: code → {original_url, state, expired}
- 命中缓存: P99 < 1ms (Redis GET)
- 未命中: 回源 MySQL，写入缓存，TTL = 10min
- 缓存更新: MQ 消费者在 update/del 事件时主动失效
```

#### 2. Gateway 每次请求新建 Proxy — 内存 & GC 风暴

**位置**: `cmd/gateway/main.go:42-51`

```go
func reverseProxy(target string) gin.HandlerFunc {
    return func(c *gin.Context) {
        remote, err := url.Parse(target)        // 每次都 Parse
        proxy := httputil.NewSingleHostReverseProxy(remote) // 每次都 new
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

**问题**: 10,000 QPS 下，每秒创建 10,000 个 `url.URL` 对象 + 10,000 个 `ReverseProxy` 对象。`ReverseProxy` 内部还维护了连接池等重结构。GC 压力巨大，STW 暂停会导致全局延迟抖动。

**影响**: Gateway 是所有流量入口，GC 抖动影响全部服务。

**优化方案**:
```go
func reverseProxy(target string) gin.HandlerFunc {
    remote, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(remote)
    return func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

---

### P1 - 严重级 (高并发下会成为瓶颈)

#### 3. GORM 未配置连接池

**位置**: `cmd/link/main.go:42-53`

```go
db0, err := gorm.Open(mysql.Open(dsn0), &gorm.Config{})
// 没有配置 MaxOpenConns, MaxIdleConns, ConnMaxLifetime
```

**问题**: Go 的 `database/sql` 默认 `MaxOpenConns = 0` (无限)，高并发下会打开成千上万的 MySQL 连接，超出 MySQL 的 `max_connections` 限制，导致 `too many connections` 错误。

**优化方案**:
```go
sqlDB, _ := db0.DB()
sqlDB.SetMaxOpenConns(100)        // 最大打开连接数
sqlDB.SetMaxIdleConns(20)         // 最大空闲连接数
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
sqlDB.SetConnMaxIdleTime(2 * time.Minute)  // 空闲连接最大存活时间
```

#### 4. 限流器内存泄漏

**位置**: `internal/common/middleware/ratelimit.go:50`

```go
limiters := &sync.Map{}
// 每个 IP 创建一个 rateBucket，永远不清理
```

**问题**: 如果有 100 万个不同 IP 访问过，`sync.Map` 中就存储了 100 万个 `rateBucket` 对象，每个约 64 字节 + mutex 开销。长时间运行后内存持续增长。

**优化方案**: 添加 TTL 清理机制，定期清理不活跃的 IP bucket。

#### 5. RabbitMQ 单 Channel

**位置**: `internal/common/mq/rabbitmq.go:17`

```go
type RabbitMQ struct {
    conn    *amqp.Connection
    channel *amqp.Channel  // 整个应用只有 1 个 channel
}
```

**问题**: AMQP 协议中，单个 Channel 的操作是串行化的。高并发 publish 场景下（比如同时创建大量短链），单 Channel 会成为吞吐瓶颈。`PublishWithContext` 带 5s 超时，意味着高负载时消息发布会排队等待。

**优化方案**: 使用 Channel 池（5-10 个 Channel），轮询分配。

#### 6. Kafka BatchTimeout 过短

**位置**: `internal/common/mq/kafka.go:22`

```go
BatchTimeout: 10,  // 10ms — 这是个 time.Duration，实际是 10 纳秒！
```

**问题**: `kafka-go` 的 `BatchTimeout` 类型是 `time.Duration`，值为 `10` 表示 10 纳秒，而非 10 毫秒。这意味着几乎每条消息都立即 flush，完全丧失批量发送的优势。在高吞吐下，每条消息都单独一次网络往返。

**优化方案**:
```go
BatchTimeout: 100 * time.Millisecond,  // 100ms 批量窗口
BatchSize:    100,                       // 最多 100 条批量发送
```

---

### P2 - 中等级 (影响效率但不致命)

#### 7. MQ 消费者单协程

**位置**: `internal/common/mq/rabbitmq.go:75-84`

```go
go func() {
    for msg := range msgs {
        if err := handler(msg.Body); err != nil {
            msg.Nack(false, true) // requeue — 可能导致毒消息循环
        } else {
            msg.Ack(false)
        }
    }
}()
```

**问题**:
- 只有 1 个 goroutine 消费，处理速度受限
- 失败消息直接 requeue，如果是数据问题会导致无限循环
- 没有死信队列处理

#### 8. MQ 发布未确认 (Publisher Confirms)

**问题**: RabbitMQ publish 没有启用 confirms 机制，消息可能丢失。在创建短链等关键操作中，如果 MQ 发送失败但已经返回成功给用户，会导致数据不一致。

#### 9. Gateway 限流 — 使用 gin.Default()

**位置**: `cmd/gateway/main.go:22`

```go
r := gin.Default()  // 带 Logger + Recovery 中间件
```

**问题**: Gateway 作为纯代理层，`gin.Default()` 的 Logger 中间件会打印每个请求的日志。10,000 QPS 下，每秒写 10,000 行日志，I/O 开销不小。应使用 `gin.New()` + 只保留 Recovery。

---

## 三、性能压测方案

### 3.1 压测工具选择

使用 **k6** (Grafana Labs)：
- Go 编写，天然高并发，可驱动万级 VU
- 自定义 metric / threshold / scenario 编排
- JSON 输出，可接 InfluxDB + Grafana 可视化

### 3.2 压测脚本清单

| 脚本 | 用途 | 场景 |
|------|------|------|
| `k6_smoke.js` | 冒烟测试，10秒预检 | 单 VU，验证服务可达 |
| `k6_redirect.js` | 核心热路径压测 | 50→200→500→1000→0 VU |
| `k6_mixed.js` | 混合负载压测 | 80%重定向 + 10%分页 + 5%创建 + 5%详情 |
| `k6_gateway.js` | Gateway 限流验证 | 3场景并行：正常/突发/多IP |
| `seed_data.py` | 测试数据填充 | 分库分表精准填充 |
| `run_bench.sh` | 一键编排脚本 | 自动化完整压测流程 |

### 3.3 场景设计

#### 场景 1: 短链重定向 (k6_redirect.js)

```
目标: GET /:shortLinkCode → 302
基准: 单实例 10,000 QPS, P99 < 50ms

阶段编排:
  1. 预热:    50 VUs,  30s  (建立连接池)
  2. 爬坡:    200 VUs, 60s
  3. 稳态:    500 VUs, 120s (模拟高峰期)
  4. 峰值:    1000 VUs, 60s (压力测试)
  5. 降压:    0 VUs,   30s  (观察恢复)

阈值:
  http_req_duration  P50<10ms  P95<30ms  P99<100ms
  http_req_failed    < 1%
  errors             < 5%
```

特点：
- `redirects: 0` — 不跟随 302，只测服务端响应时间
- 动态短链码池 (与 seed_data.py 输出联动)
- 支持 `X-Cache` header 缓存命中率追踪

#### 场景 2: 混合负载 (k6_mixed.js)

```
流量分布 (模拟真实场景):
  80%  短链重定向  GET /:code         无认证
  10%  分页查询    POST /link/v1/page  需 token
  5%   创建短链    POST /link/v1/add   需 token
  5%   短链详情    POST /link/v1/detail 需 token

阶段编排:
  warm(100) → ramp(300) → steady(500, 3min) → peak(800) → cool(0)

阈值:
  redirect_duration  P99<50ms
  api_duration       P99<300ms
  http_req_failed    < 2%
```

特点：
- 自动登录获取 JWT token (支持 `LOGIN_PHONE` / `LOGIN_PWD` 环境变量)
- 按概率分配流量，模拟真实业务分布
- 分别追踪 redirect 和 API 延迟

#### 场景 3: Gateway 限流验证 (k6_gateway.js)

```
3 个并行场景:
  1. normal_flow  — 50 VU 持续 60s (不应被限流)
  2. burst_flow   — 50→300 req/s 突发 (应触发 429)
  3. multi_ip     — 100 VU + X-Forwarded-For (独立限流)

阈值:
  rate_limited{burst}  > 10%   (突发场景验证限流生效)
  gateway_latency      P99<200ms
```

### 3.4 压测前置准备

1. **测试数据**: `python3 test/stress/seed_data.py --count 10000`
   - 精准匹配分库分表路由: db_prefix ∈ {0,1,a}, table_suffix ∈ {0,a}
   - `INSERT IGNORE` 幂等，可重复执行
   - 输出样本 codes 自动注入 k6 脚本

2. **监控建议** (可选):
   - Go runtime: goroutine 数、GC 暂停、内存
   - MySQL: 连接数、慢查询、buffer pool hit rate
   - Redis: 命中率、内存、OPS
   - RabbitMQ: 队列深度、消费速率、unacked
   - Kafka: produce/consume lag

### 3.5 一键执行

```bash
# 全部场景
./test/stress/run_bench.sh

# 仅重定向场景
./test/stress/run_bench.sh redirect --count 50000

# 自定义参数
./test/stress/run_bench.sh mixed --base-url http://prod:8888 --token eyJ...

# 跳过 Docker (已有环境)
./test/stress/run_bench.sh all --skip-docker

# 冒烟测试 (10 秒)
k6 run test/stress/k6_smoke.js
```

---

## 四、优化实施方案

### 阶段一: 立即修复 (1-2 天) — 效果最大

#### 4.1 热路径添加 Redis 缓存

**文件**: `internal/link/controller/link_api.go`

```go
// 新增 Redis 缓存层
const shortLinkCachePrefix = "sl:redirect:"
const shortLinkCacheTTL = 10 * time.Minute

type cachedShortLink struct {
    OriginalUrl string `json:"original_url"`
    State       string `json:"state"`
    Del         int    `json:"del"`
    Expired     *string `json:"expired"`
}

func (ctrl *LinkApiController) Dispatch(c *gin.Context) {
    code := c.Param("shortLinkCode")
    if code == "" || !ctrl.codeRegexp.MatchString(code) {
        c.String(http.StatusBadRequest, "invalid short link code")
        return
    }

    // Step 1: 尝试从 Redis 缓存获取
    cacheKey := shortLinkCachePrefix + code
    cached, err := ctrl.rdb.Get(c, cacheKey).Result()
    if err == nil {
        var sl cachedShortLink
        json.Unmarshal([]byte(cached), &sl)
        if sl.Del == 1 || sl.State == "LOCK" {
            c.String(http.StatusForbidden, "short link is locked or deleted")
            return
        }
        // 缓存命中，直接重定向
        ctrl.sendVisitLog(c, code)
        c.Redirect(http.StatusFound, util.RemoveUrlPrefix(sl.OriginalUrl))
        return
    }

    // Step 2: 缓存未命中，回源 MySQL
    dbPrefix, tableSuffix := sharding.RouteShortLink(code)
    dbIdx := sl.getDBIndexByPrefix(dbPrefix)
    tableName := sharding.GetTableName("short_link", tableSuffix)

    var shortLink cachedShortLink
    err = ctrl.dbs[dbIdx].Table(tableName).
        Where("code = ? AND del = 0", code).
        First(&shortLink).Error
    if err != nil {
        c.String(http.StatusNotFound, "short link not found")
        return
    }

    // Step 3: 写入缓存
    data, _ := json.Marshal(shortLink)
    ctrl.rdb.Set(c, cacheKey, data, shortLinkCacheTTL)

    if shortLink.Del == 1 || shortLink.State == "LOCK" {
        c.String(http.StatusForbidden, "short link is locked or deleted")
        return
    }

    ctrl.sendVisitLog(c, code)
    c.Redirect(http.StatusFound, util.RemoveUrlPrefix(shortLink.OriginalUrl))
}
```

**缓存失效**: MQ 消费者处理 update/del 事件时，删除对应 Redis key:
```go
ctrl.rdb.Del(ctx, "sl:redirect:"+code)
```

#### 4.2 Gateway 复用 Proxy 实例

**文件**: `cmd/gateway/main.go`

```go
func reverseProxy(target string) gin.HandlerFunc {
    remote, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(remote)
    return func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

#### 4.3 GORM 连接池配置

**文件**: `cmd/link/main.go` (及其他服务的 main.go)

```go
func configureDB(db *gorm.DB) {
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetMaxIdleConns(20)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)
    sqlDB.SetConnMaxIdleTime(2 * time.Minute)
}
```

### 阶段二: 短期优化 (1 周)

#### 4.4 限流器添加 TTL 清理

```go
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
    limiters := &sync.Map{}
    // 定期清理不活跃 bucket
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        for range ticker.C {
            limiters.Range(func(key, value interface{}) bool {
                bucket := value.(*rateBucket)
                bucket.mu.Lock()
                if time.Since(bucket.lastRefill) > 10*time.Minute {
                    limiters.Delete(key)
                }
                bucket.mu.Unlock()
                return true
            })
        }
    }()
    // ... rest unchanged
}
```

#### 4.5 Kafka BatchTimeout 修复

```go
w := &kafka.Writer{
    Addr:         kafka.TCP(brokers...),
    Topic:        topic,
    Balancer:     &kafka.LeastBytes{},
    BatchTimeout: 100 * time.Millisecond,  // 修复: 100ms
    BatchSize:    100,
    RequiredAcks: kafka.RequireOne,
}
```

#### 4.6 RabbitMQ Channel 池化

```go
type RabbitMQ struct {
    conn     *amqp.Connection
    channels []*amqp.Channel
    idx      uint64
}

func (r *RabbitMQ) getChannel() *amqp.Channel {
    idx := atomic.AddUint64(&r.idx, 1)
    return r.channels[idx%uint64(len(r.channels))]
}
```

### 阶段三: 中期优化 (2-4 周)

#### 4.7 Gateway 性能优化

- 使用 `gin.New()` 替代 `gin.Default()`，去掉默认 Logger
- 考虑使用 `fasthttp` 替代标准 `net/http` 的反向代理（吞吐提升 2-3x）
- 添加连接复用配置

#### 4.8 MQ 消费者并发化

```go
func (r *RabbitMQ) Consume(queueName, consumerTag string, concurrency int, handler func([]byte) error) error {
    msgs, err := r.channel.Consume(queueName, consumerTag, false, false, false, false, nil)
    if err != nil {
        return err
    }
    for i := 0; i < concurrency; i++ {
        go func(workerID int) {
            for msg := range msgs {
                if err := handler(msg.Body); err != nil {
                    log.Printf("[MQ] worker %d error on %s: %v", workerID, queueName, err)
                    msg.Nack(false, false) // 不 requeue，进死信
                } else {
                    msg.Ack(false)
                }
            }
        }(i)
    }
    return nil
}
```

#### 4.9 添加 Benchmark 测试

**文件**: `internal/link/controller/link_api_test.go`

```go
func BenchmarkDispatch(b *testing.B) {
    // Setup: mock DB, Redis, Kafka
    ctrl := setupBenchmarkController()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            c.Request = httptest.NewRequest("GET", "/abc123", nil)
            c.Params = gin.Param{{Key: "shortLinkCode", Value: "abc123"}}
            ctrl.Dispatch(c)
        }
    })
}
```

---

## 五、优化效果预估

| 优化项 | 当前瓶颈 | 优化后预期 | 影响 QPS |
|--------|----------|-----------|---------|
| Redis 缓存短链 | MySQL P99 ~50ms | Redis P99 <1ms | **50x** |
| Proxy 复用 | 每请求 new 对象 | 零分配 | 3-5x |
| GORM 连接池 | 无限连接/连接耗尽 | 稳定 100 连接 | 稳定性 |
| Kafka 超时修复 | 10ns flush | 100ms 批量 | 10x 吞吐 |
| 限流器清理 | 内存泄漏 | O(1) 清理 | 运维安全 |

**预估**: 优化后单实例短链重定向 QPS 可从 ~2,000 提升到 **20,000+**。

---

## 六、下一步行动

1. [ ] 实施 Redis 缓存 (P0，效果最大)
2. [ ] 修复 Gateway Proxy 复用 (P0，改 3 行代码)
3. [ ] 配置 GORM 连接池 (P1)
4. [ ] 编写 k6 压测脚本并执行基准测试
5. [ ] 修复 Kafka BatchTimeout (P1)
6. [ ] 限流器 TTL 清理 (P2)
7. [ ] RabbitMQ Channel 池化 (P2)
