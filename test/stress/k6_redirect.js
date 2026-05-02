/**
 * k6 压测脚本 — 短链重定向 (核心热路径)
 *
 * 使用方法:
 *   1. 确保服务已启动 (docker-compose up)
 *   2. 安装 k6: brew install k6
 *   3. 预填充测试数据 (见下方说明)
 *   4. 运行: k6 run test/stress/k6_redirect.js
 *   5. 带 InfluxDB 输出: k6 run --out influxdb=http://localhost:8086/k6 test/stress/k6_redirect.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');
const redirectDuration = new Trend('redirect_duration');
const cacheHitRate = new Rate('cache_hits');
const totalRequests = new Counter('total_requests');

// 测试配置
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';

// 预生成的短链码列表 (需要提前在数据库中插入测试数据)
// 运行前请确保数据库中有这些短链记录
const SHORT_CODES = [
  '0abc0', '0def0', '0ghi0', '0jkl0', '0mno0',
  '0pqr0', '0stu0', '0vwx0', '0yza0', '0bcd0',
  '1abc1', '1def1', '1ghi1', '1jkl1', '1mno1',
  '1pqr1', '1stu1', '1vwx1', '1yza1', '1bcd1',
  'aabc0', 'adef0', 'aghi0', 'ajkl0', 'amno0',
];

// 随机选择一个短链码
function randomCode() {
  return SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
}

export const options = {
  scenarios: {
    // 场景 1: 短链重定向压测
    redirect_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },    // 预热
        { duration: '60s', target: 200 },   // 爬坡
        { duration: '120s', target: 500 },  // 稳态 (模拟高峰)
        { duration: '60s', target: 1000 },  // 峰值压力
        { duration: '30s', target: 0 },     // 降压
      ],
      gracefulRampDown: '10s',
    },
  },

  thresholds: {
    // 性能阈值 (超过则测试失败)
    'http_req_duration': [
      'p(50)<10',    // P50 < 10ms
      'p(95)<30',    // P95 < 30ms
      'p(99)<100',   // P99 < 100ms
    ],
    'http_req_failed': ['rate<0.01'],  // 错误率 < 1%
    'errors': ['rate<0.01'],
    'iterations': ['rate>3000'],       // QPS > 3000
  },
};

export default function () {
  const code = randomCode();
  const url = `${BASE_URL}/${code}`;

  const res = http.get(url, {
    redirects: 0,  // 不自动跟随重定向，我们关心的是 302 响应时间
    timeout: '10s',
    tags: { name: 'Redirect' },
  });

  totalRequests.add(1);

  // 检查响应
  const isRedirect = res.status === 302;
  const isNotFound = res.status === 404;
  const isForbidden = res.status === 403;
  const isOk = isRedirect || isNotFound || isForbidden;

  check(res, {
    'status is 302/404/403': (r) => r.status === 302 || r.status === 404 || r.status === 403,
    'response time < 100ms': (r) => r.timings.duration < 100,
    'response time < 50ms': (r) => r.timings.duration < 50,
  });

  errorRate.add(!isOk);
  redirectDuration.add(res.timings.duration);

  // 检查是否有缓存命中的迹象 (X-Cache header，需要后端添加)
  if (res.headers['X-Cache']) {
    cacheHitRate.add(res.headers['X-Cache'] === 'HIT');
  }
}

// 测试生命周期钩子
export function setup() {
  // 预检：确保服务可达
  const res = http.get(`${BASE_URL}/health`, { timeout: '5s' });
  if (res.status !== 200) {
    console.warn(`Warning: Service health check returned ${res.status}`);
  }
  console.log(`Target: ${BASE_URL}`);
  console.log(`Test codes: ${SHORT_CODES.length} short link codes`);
}

export function teardown(data) {
  console.log('=== Stress test completed ===');
}
