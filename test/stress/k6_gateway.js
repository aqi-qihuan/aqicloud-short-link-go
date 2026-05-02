/**
 * k6 压测脚本 — Gateway 限流验证 + 代理性能
 *
 * 验证目标:
 *   1. Gateway 限流器在高并发下的正确性 (100 req/s/IP, burst 200)
 *   2. Gateway 反向代理的延迟开销
 *   3. 限流触发后返回 429 的正确性
 *
 * 使用方法:
 *   k6 run test/stress/k6_gateway.js
 *   k6 run -e BASE_URL=http://prod-gateway:8888 test/stress/k6_gateway.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ──────────────────────────────────────
// 自定义指标
// ──────────────────────────────────────
const errorRate     = new Rate('errors');
const gatewayLatency = new Trend('gateway_latency');
const rateLimited   = new Rate('rate_limited');
const proxyOverhead = new Trend('proxy_overhead_ms');

// ──────────────────────────────────────
// 配置
// ──────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';
const LINK_DIRECT = __ENV.LINK_DIRECT || 'http://localhost:8003'; // 直连 link 服务

const DB_PREFIXES  = ['0', '1', 'a'];
const TABLE_SUFFIX = ['0', 'a'];

function base62Encode(num) {
  const chars = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
  if (num === 0) return chars[0];
  let r = '';
  while (num > 0) { r = chars[num % 62] + r; num = Math.floor(num / 62); }
  return r;
}

const SHORT_CODES = [];
for (let i = 0; i < 2000; i++) {
  const hashVal = Math.floor(Math.random() * 4294967295);
  let code = base62Encode(hashVal);
  while (code.length < 5) code = '0' + code;
  SHORT_CODES.push(DB_PREFIXES[i % 3] + code.substring(0, 5) + TABLE_SUFFIX[i % 2]);
}

function randomCode() {
  return SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
}

// ──────────────────────────────────────
// 场景: 3 个独立场景并行
// ──────────────────────────────────────
export const options = {
  scenarios: {
    // 场景 1: 正常流量 — 不应被限流
    normal_flow: {
      executor: 'constant-vus',
      vus: 50,
      duration: '60s',
      tags: { scenario: 'normal' },
    },

    // 场景 2: 突发流量 — 触发限流
    burst_flow: {
      executor: 'ramping-arrival-rate',
      startRate: 50,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 500,
      stages: [
        { duration: '10s', target: 50 },   // 正常
        { duration: '10s', target: 300 },  // 突发 — 超过 100r/s 限流
        { duration: '20s', target: 300 },  // 持续突发
        { duration: '10s', target: 50 },   // 恢复
        { duration: '10s', target: 50 },   // 观察恢复
      ],
      exec: 'burstTest',
      tags: { scenario: 'burst' },
    },

    // 场景 3: 多 IP 模拟 — 不同 IP 应独立限流
    multi_ip: {
      executor: 'constant-vus',
      vus: 100,
      duration: '60s',
      exec: 'multiIPTest',
      tags: { scenario: 'multi_ip' },
    },
  },

  thresholds: {
    'http_req_duration':            ['p(99)<500'],
    'http_req_failed':              ['rate<0.05'],
    'gateway_latency':              ['p(99)<200'],
    'errors':                       ['rate<0.05'],
    'rate_limited{scenario:burst}': ['rate>0.1'],   // 突发场景至少 10% 被限流
  },
};

// ──────────────────────────────────────
// 场景 1: 正常流量
// ──────────────────────────────────────
export default function () {
  const code = randomCode();
  const res = http.get(`${BASE_URL}/${code}`, {
    redirects: 0,
    timeout: '10s',
    tags: { name: 'normal_flow' },
  });

  check(res, {
    'not rate-limited':  (r) => r.status !== 429,
    'status ok':         (r) => [302, 404, 403].includes(r.status) || r.status === 429,
    'latency < 200ms':   (r) => r.timings.duration < 200,
  });

  gatewayLatency.add(res.timings.duration);
  errorRate.add(res.status === 0 || (res.status !== 302 && res.status !== 404 && res.status !== 403 && res.status !== 429));
}

// ──────────────────────────────────────
// 场景 2: 突发流量 — 验证限流触发
// ──────────────────────────────────────
export function burstTest() {
  const code = randomCode();
  const res = http.get(`${BASE_URL}/${code}`, {
    redirects: 0,
    timeout: '10s',
    tags: { name: 'burst_flow' },
  });

  const isRateLimited = res.status === 429;

  check(res, {
    'response received':     (r) => r.status !== 0,
    'valid status':          (r) => [302, 404, 403, 429].includes(r.status),
    '429 has rate msg':      (r) => !isRateLimited || (r.body && r.body.includes('rate limit')),
  });

  rateLimited.add(isRateLimited);
  gatewayLatency.add(res.timings.duration);
  errorRate.add(res.status === 0);
}

// ──────────────────────────────────────
// 场景 3: 多 IP 模拟
// ──────────────────────────────────────
export function multiIPTest() {
  const code = randomCode();

  // 通过 X-Forwarded-For 模拟不同客户端 IP
  // 注意: Gateway 可能不信任此 header，取决于部署配置
  const fakeIP = `${10 + Math.floor(Math.random() * 240)}.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}.${1 + Math.floor(Math.random() * 254)}`;

  const res = http.get(`${BASE_URL}/${code}`, {
    redirects: 0,
    timeout: '10s',
    headers: { 'X-Forwarded-For': fakeIP },
    tags: { name: 'multi_ip_flow' },
  });

  check(res, {
    'response received':  (r) => r.status !== 0,
    'valid status':       (r) => [302, 404, 403, 429].includes(r.status),
    'latency < 200ms':    (r) => r.timings.duration < 200,
  });

  gatewayLatency.add(res.timings.duration);
  errorRate.add(res.status === 0);
}

// ──────────────────────────────────────
// 生命周期钩子
// ──────────────────────────────────────
export function setup() {
  console.log('========================================');
  console.log('  Gateway Rate-Limit & Proxy Stress');
  console.log('========================================');
  console.log(`Gateway:    ${BASE_URL}`);
  console.log(`Link直连:   ${LINK_DIRECT}`);
  console.log('');

  const gwRes = http.get(`${BASE_URL}/${SHORT_CODES[0]}`, { redirects: 0, timeout: '5s' });
  console.log(`Gateway pre-check:  HTTP ${gwRes.status} (${gwRes.timings.duration.toFixed(1)}ms)`);

  const linkRes = http.get(`${LINK_DIRECT}/${SHORT_CODES[0]}`, { redirects: 0, timeout: '5s' });
  console.log(`Link直连 pre-check: HTTP ${linkRes.status} (${linkRes.timings.duration.toFixed(1)}ms)`);

  if (gwRes.status === 0) {
    console.error('ERROR: Gateway 不可达');
  }
  console.log('');

  return { startTime: Date.now() };
}

export function teardown(data) {
  const duration = ((Date.now() - data.startTime) / 1000).toFixed(1);
  console.log(`Gateway stress test completed in ${duration}s`);
}
