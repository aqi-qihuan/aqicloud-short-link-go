/**
 * k6 压测脚本 — 短链重定向 (核心热路径)
 *
 * 使用方法:
 *   1. 确保服务已启动 (docker-compose up -d)
 *   2. 填充数据: python3 test/stress/seed_data.py --count 10000
 *   3. 安装 k6: brew install k6
 *   4. 运行: k6 run test/stress/k6_redirect.js
 *   5. 自定义目标: k6 run -e BASE_URL=http://example.com test/stress/k6_redirect.js
 *   6. 带 InfluxDB: k6 run --out influxdb=http://localhost:8086/k6 test/stress/k6_redirect.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ──────────────────────────────────────
// 自定义指标
// ──────────────────────────────────────
const errorRate     = new Rate('errors');
const redirectDur   = new Trend('redirect_duration');
const cacheHitRate  = new Rate('cache_hits');
const totalRequests = new Counter('total_requests');
const notFoundRate  = new Rate('not_found_rate');

// ──────────────────────────────────────
// 配置
// ──────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';

// 分库分表路由常量 (与项目一致)
const DB_PREFIXES  = ['0', '1', 'a'];
const TABLE_SUFFIX = ['0', 'a'];

// 动态生成短链码 (base62 编码 + 分库前缀 + 分表后缀)
function base62Encode(num) {
  const chars = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  if (num === 0) return chars[0];
  let r = '';
  while (num > 0) { r = chars[num % 62] + r; num = Math.floor(num / 62); }
  return r;
}

// 预生成码池 (脚本启动时一次性生成)
const SHORT_CODES = [];
const CODE_POOL_SIZE = parseInt(__ENV.CODE_POOL_SIZE || '10000');
const SEED_FROM_DB   = __ENV.SHORT_CODES || ''; // 从 seed_data.py 输出注入

if (SEED_FROM_DB) {
  // 使用 seed_data.py 输出的真实短链码
  SEED_FROM_DB.split(',').forEach(c => {
    const trimmed = c.trim().replace(/[\[\]'\s]/g, '');
    if (trimmed) SHORT_CODES.push(trimmed);
  });
  console.log(`从环境变量加载 ${SHORT_CODES.length} 个短链码`);
} else {
  // 动态生成 (仅用于无数据库环境的快速测试)
  for (let i = 0; i < CODE_POOL_SIZE; i++) {
    const hashVal = Math.floor(Math.random() * 4294967295);
    let code = base62Encode(hashVal);
    while (code.length < 5) code = '0' + code;
    code = code.substring(0, 5);
    const prefix = DB_PREFIXES[i % DB_PREFIXES.length];
    const suffix = TABLE_SUFFIX[i % TABLE_SUFFIX.length];
    SHORT_CODES.push(prefix + code + suffix);
  }
  console.log(`动态生成 ${SHORT_CODES.length} 个短链码 (注意: 需要先填充数据库才有 302 响应)`);
}

function randomCode() {
  return SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
}

// ──────────────────────────────────────
// 压测场景配置
// ──────────────────────────────────────
export const options = {
  scenarios: {
    redirect_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50  },   // 预热 — 建立连接池
        { duration: '60s', target: 200 },   // 爬坡
        { duration: '120s', target: 500 },  // 稳态 (模拟高峰期)
        { duration: '60s', target: 1000 },  // 峰值压力
        { duration: '30s', target: 0   },   // 降压
      ],
      gracefulRampDown: '10s',
    },
  },

  thresholds: {
    'http_req_duration': [
      'p(50)<10',    // P50 < 10ms
      'p(95)<30',    // P95 < 30ms
      'p(99)<100',   // P99 < 100ms
    ],
    'http_req_failed': ['rate<0.01'],  // 连接失败率 < 1%
    'errors':          ['rate<0.05'],  // 业务错误率 < 5%
    'redirect_duration': ['p(99)<100'],
  },
};

// ──────────────────────────────────────
// 主测试函数
// ──────────────────────────────────────
export default function () {
  const code = randomCode();
  const url = `${BASE_URL}/${code}`;

  const res = http.get(url, {
    redirects: 0,         // 不跟随重定向 — 我们只关心 302 响应时间
    timeout: '10s',
    tags: { name: 'GET /:code' },
  });

  totalRequests.add(1);

  const isRedirect  = res.status === 302;
  const isNotFound  = res.status === 404;
  const isForbidden = res.status === 403;
  const isOk = isRedirect || isNotFound || isForbidden;

  check(res, {
    'status is 302/404/403':  (r) => [302, 404, 403].includes(r.status),
    'response time < 100ms':  (r) => r.timings.duration < 100,
    'response time < 50ms':   (r) => r.timings.duration < 50,
    'response time < 10ms':   (r) => r.timings.duration < 10,
  });

  errorRate.add(!isOk);
  notFoundRate.add(isNotFound);
  redirectDur.add(res.timings.duration);

  // 检查缓存命中 (后端需添加 X-Cache header)
  const xcache = res.headers['X-Cache'] || res.headers['x-cache'];
  if (xcache) {
    cacheHitRate.add(xcache === 'HIT');
  }
}

// ──────────────────────────────────────
// 生命周期钩子
// ──────────────────────────────────────
export function setup() {
  console.log('========================================');
  console.log('  AqiCloud Short-Link Redirect Stress');
  console.log('========================================');
  console.log(`Target:       ${BASE_URL}`);
  console.log(`Code pool:    ${SHORT_CODES.length} codes`);
  console.log(`Stages:       warm(50) → ramp(200) → steady(500) → peak(1000) → cool(0)`);
  console.log(`Thresholds:   P50<10ms, P95<30ms, P99<100ms`);
  console.log('');

  // 预检: 确保 Gateway 可达
  const res = http.get(`${BASE_URL}/${SHORT_CODES[0]}`, {
    redirects: 0, timeout: '5s',
  });
  console.log(`Pre-check:    ${BASE_URL}/${SHORT_CODES[0]} → HTTP ${res.status} (${res.timings.duration.toFixed(1)}ms)`);

  if (res.status === 0) {
    console.error('ERROR: Gateway 不可达，请检查 docker-compose up');
  } else if (res.status === 404) {
    console.warn('WARN: 返回 404 — 请先运行 seed_data.py 填充测试数据');
  }

  return { startTime: Date.now() };
}

export function teardown(data) {
  const duration = ((Date.now() - data.startTime) / 1000).toFixed(1);
  console.log('');
  console.log(`========================================`);
  console.log(`  Completed in ${duration}s`);
  console.log(`  Total requests: ${totalRequests.value || 'check summary'}`);
  console.log(`========================================`);
}
