/**
 * k6 压测脚本 — 混合负载测试 (模拟真实流量分布)
 *
 * 流量分布:
 *   - 短链重定向: 80% (无认证)
 *   - 分页查询:   10% (需认证)
 *   - 短链创建:   5%  (需认证)
 *   - 短链详情:   5%  (需认证)
 *
 * 使用方法:
 *   k6 run test/stress/k6_mixed.js
 *   k6 run -e AUTH_TOKEN=xxx test/stress/k6_mixed.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');
const redirectP99 = new Trend('redirect_p99');
const apiP99 = new Trend('api_p99');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';

const SHORT_CODES = [
  '0abc0', '0def0', '0ghi0', '0jkl0', '0mno0',
  '0pqr0', '0stu0', '0vwx0', '0yza0', '0bcd0',
  '1abc1', '1def1', '1ghi1', '1jkl1', '1mno1',
  'aabc0', 'adef0', 'aghi0', 'ajkl0', 'amno0',
];

function randomCode() {
  return SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
}

export const options = {
  scenarios: {
    // 主场景：混合流量
    mixed_traffic: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 },   // 预热
        { duration: '60s', target: 300 },   // 爬坡
        { duration: '180s', target: 500 },  // 高峰稳态 3 分钟
        { duration: '60s', target: 800 },   // 峰值冲击
        { duration: '60s', target: 300 },   // 回落
        { duration: '30s', target: 0 },     // 降压
      ],
      gracefulRampDown: '15s',
    },
  },

  thresholds: {
    'http_req_duration{scenario:redirect}': ['p(99)<50'],
    'http_req_duration{scenario:api}': ['p(99)<200'],
    'http_req_failed': ['rate<0.02'],
    'errors': ['rate<0.02'],
  },
};

export default function () {
  // 按概率分配流量
  const rand = Math.random();

  if (rand < 0.80) {
    // 80% — 短链重定向
    testRedirect();
  } else if (rand < 0.90) {
    // 10% — 分页查询
    testPageQuery();
  } else if (rand < 0.95) {
    // 5% — 创建短链
    testCreateLink();
  } else {
    // 5% — 短链详情
    testLinkDetail();
  }
}

function testRedirect() {
  const code = randomCode();
  const res = http.get(`${BASE_URL}/${code}`, {
    redirects: 0,
    timeout: '10s',
    tags: { scenario: 'redirect', name: 'GET /:code' },
  });

  check(res, {
    'redirect status ok': (r) => [302, 404, 403].includes(r.status),
    'redirect < 50ms': (r) => r.timings.duration < 50,
  });

  errorRate.add(![302, 404, 403].includes(res.status));
  redirectP99.add(res.timings.duration);
}

function testPageQuery() {
  if (!AUTH_TOKEN) return;

  const payload = JSON.stringify({
    group_id: 1,
    page: 1,
    size: 20,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/page`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
    timeout: '10s',
    tags: { scenario: 'api', name: 'POST /page' },
  });

  check(res, {
    'page query ok': (r) => r.status === 200,
    'page query < 200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(res.status !== 200);
  apiP99.add(res.timings.duration);
}

function testCreateLink() {
  if (!AUTH_TOKEN) return;

  const uniqueId = Date.now() + Math.random().toString(36).substring(7);
  const payload = JSON.stringify({
    domain_type: 'OFFICIAL',
    original_url: `https://example.com/perf-test/${uniqueId}`,
    title: `Perf Test ${uniqueId}`,
    group_id: 1,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/add`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
    timeout: '15s',
    tags: { scenario: 'api', name: 'POST /add' },
  });

  check(res, {
    'create link ok': (r) => r.status === 200,
    'create link < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(res.status !== 200);
  apiP99.add(res.timings.duration);
}

function testLinkDetail() {
  if (!AUTH_TOKEN) return;

  const payload = JSON.stringify({
    group_id: 1,
    mapping_id: 1,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/detail`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${AUTH_TOKEN}`,
    },
    timeout: '10s',
    tags: { scenario: 'api', name: 'POST /detail' },
  });

  check(res, {
    'detail status ok': (r) => r.status === 200 || r.status === 404,
    'detail < 100ms': (r) => r.timings.duration < 100,
  });

  errorRate.add(res.status !== 200 && res.status !== 404);
  apiP99.add(res.timings.duration);
}

export function setup() {
  console.log(`=== Mixed Load Test ===`);
  console.log(`Target: ${BASE_URL}`);
  console.log(`Auth Token: ${AUTH_TOKEN ? 'provided' : 'NOT provided (API tests will be skipped)'}`);
  console.log(`Traffic distribution: 80% redirect, 10% page, 5% create, 5% detail`);
}
