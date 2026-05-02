/**
 * k6 压测脚本 — 混合负载测试 (模拟真实流量分布)
 *
 * 流量分布 (模拟短链接系统真实场景):
 *   - 短链重定向: 80%  (GET /:code — 无认证，热路径)
 *   - 分页查询:   10%  (POST /link-server/api/link/v1/page — 需 token)
 *   - 短链创建:   5%   (POST /link-server/api/link/v1/add — 需 token)
 *   - 短链详情:   5%   (POST /link-server/api/link/v1/detail — 需 token)
 *
 * 使用方法:
 *   # 自动登录获取 token (需要先注册用户)
 *   k6 run -e LOGIN_PHONE=13800000001 -e LOGIN_PWD=123456 test/stress/k6_mixed.js
 *
 *   # 手动指定 token
 *   k6 run -e AUTH_TOKEN=eyJhbGciOiJIUz... test/stress/k6_mixed.js
 *
 *   # 自定义参数
 *   k6 run -e BASE_URL=http://prod.example.com -e AUTH_TOKEN=xxx test/stress/k6_mixed.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ──────────────────────────────────────
// 自定义指标
// ──────────────────────────────────────
const errorRate    = new Rate('errors');
const redirectP99  = new Trend('redirect_duration');
const apiP99       = new Trend('api_duration');
const redirectQPS  = new Counter('redirect_requests');
const apiQPS       = new Counter('api_requests');

// ──────────────────────────────────────
// 配置
// ──────────────────────────────────────
const BASE_URL    = __ENV.BASE_URL    || 'http://localhost:8888';
const AUTH_TOKEN  = __ENV.AUTH_TOKEN  || '';
const LOGIN_PHONE = __ENV.LOGIN_PHONE || '';
const LOGIN_PWD   = __ENV.LOGIN_PWD   || '';

// 短链码 (seed_data.py 输出)
const SEED_FROM_DB = __ENV.SHORT_CODES || '';
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
if (SEED_FROM_DB) {
  SEED_FROM_DB.split(',').forEach(c => {
    const trimmed = c.trim().replace(/[\[\]'\s]/g, '');
    if (trimmed) SHORT_CODES.push(trimmed);
  });
} else {
  for (let i = 0; i < 5000; i++) {
    const hashVal = Math.floor(Math.random() * 4294967295);
    let code = base62Encode(hashVal);
    while (code.length < 5) code = '0' + code;
    code = code.substring(0, 5);
    SHORT_CODES.push(DB_PREFIXES[i % 3] + code + TABLE_SUFFIX[i % 2]);
  }
}

function randomCode() {
  return SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
}

// ──────────────────────────────────────
// 登录获取 token (setup 阶段自动执行)
// ──────────────────────────────────────
function login(phone, pwd) {
  const payload = JSON.stringify({ phone, pwd });
  const res = http.post(`${BASE_URL}/account-server/api/account/v1/login`, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '10s',
  });

  if (res.status === 200) {
    const body = JSON.parse(res.body);
    if (body.code === 0 && body.data) {
      return body.data; // JWT token
    }
  }
  return null;
}

// ──────────────────────────────────────
// 压测场景配置
// ──────────────────────────────────────
export const options = {
  scenarios: {
    mixed_traffic: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s',  target: 100 },  // 预热
        { duration: '60s',  target: 300 },  // 爬坡
        { duration: '180s', target: 500 },  // 高峰稳态 3 分钟
        { duration: '60s',  target: 800 },  // 峰值冲击
        { duration: '60s',  target: 300 },  // 回落
        { duration: '30s',  target: 0   },  // 降压
      ],
      gracefulRampDown: '15s',
    },
  },

  thresholds: {
    'http_req_duration':  ['p(99)<200'],
    'http_req_failed':    ['rate<0.02'],
    'errors':             ['rate<0.05'],
    'redirect_duration':  ['p(99)<50'],
    'api_duration':       ['p(99)<300'],
  },
};

// ──────────────────────────────────────
// 主测试函数 — 按概率分配流量
// ──────────────────────────────────────
let token = AUTH_TOKEN;

export default function () {
  const rand = Math.random();

  if (rand < 0.80) {
    testRedirect();           // 80% — 短链重定向
  } else if (rand < 0.90) {
    testPageQuery();          // 10% — 分页查询
  } else if (rand < 0.95) {
    testCreateLink();         // 5%  — 创建短链
  } else {
    testLinkDetail();         // 5%  — 短链详情
  }
}

// ── 80% 短链重定向 ──
function testRedirect() {
  const code = randomCode();
  const res = http.get(`${BASE_URL}/${code}`, {
    redirects: 0,
    timeout: '10s',
    tags: { scenario: 'redirect', name: 'GET /:code' },
  });

  check(res, {
    'redirect status ok':  (r) => [302, 404, 403].includes(r.status),
    'redirect < 50ms':     (r) => r.timings.duration < 50,
  });

  errorRate.add(![302, 404, 403].includes(res.status));
  redirectP99.add(res.timings.duration);
  redirectQPS.add(1);
}

// ── 10% 分页查询 ──
function testPageQuery() {
  if (!token) return;

  const payload = JSON.stringify({
    groupId: 1,
    page: 1,
    size: 20,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/page`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'token': token,
    },
    timeout: '10s',
    tags: { scenario: 'api', name: 'POST /page' },
  });

  check(res, {
    'page query ok':    (r) => r.status === 200,
    'page query < 200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(res.status !== 200);
  apiP99.add(res.timings.duration);
  apiQPS.add(1);
}

// ── 5% 创建短链 ──
function testCreateLink() {
  if (!token) return;

  const uniqueId = Date.now() + Math.random().toString(36).substring(7);
  const payload = JSON.stringify({
    domainType: 'OFFICIAL',
    originalUrl: `https://example.com/perf-test/${uniqueId}`,
    title: `Perf Test ${uniqueId}`,
    groupId: 1,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/add`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'token': token,
    },
    timeout: '15s',
    tags: { scenario: 'api', name: 'POST /add' },
  });

  check(res, {
    'create link ok':     (r) => r.status === 200,
    'create link < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(res.status !== 200);
  apiP99.add(res.timings.duration);
  apiQPS.add(1);
}

// ── 5% 短链详情 ──
function testLinkDetail() {
  if (!token) return;

  const payload = JSON.stringify({
    groupId: 1,
    mappingId: 1,
  });

  const res = http.post(`${BASE_URL}/link-server/api/link/v1/detail`, payload, {
    headers: {
      'Content-Type': 'application/json',
      'token': token,
    },
    timeout: '10s',
    tags: { scenario: 'api', name: 'POST /detail' },
  });

  check(res, {
    'detail status ok': (r) => r.status === 200 || r.status === 404,
    'detail < 100ms':   (r) => r.timings.duration < 100,
  });

  errorRate.add(res.status !== 200 && res.status !== 404);
  apiP99.add(res.timings.duration);
  apiQPS.add(1);
}

// ──────────────────────────────────────
// 生命周期钩子
// ──────────────────────────────────────
export function setup() {
  console.log('========================================');
  console.log('  AqiCloud Mixed Load Stress Test');
  console.log('========================================');
  console.log(`Target:     ${BASE_URL}`);
  console.log(`Code pool:  ${SHORT_CODES.length} codes`);
  console.log('');

  // 自动登录获取 token
  if (!AUTH_TOKEN && LOGIN_PHONE && LOGIN_PWD) {
    console.log(`Auto-login: ${LOGIN_PHONE}...`);
    const t = login(LOGIN_PHONE, LOGIN_PWD);
    if (t) {
      token = t;
      console.log(`Token:      obtained (${t.substring(0, 20)}...)`);
    } else {
      console.error('ERROR: 登录失败！请检查 LOGIN_PHONE / LOGIN_PWD');
      console.warn('WARN: 需认证的 API 将被跳过，仅测试重定向');
    }
  } else if (AUTH_TOKEN) {
    token = AUTH_TOKEN;
    console.log(`Token:      from env (${token.substring(0, 20)}...)`);
  } else {
    console.warn('WARN: 未提供 token，仅测试短链重定向 (80% 流量)');
  }

  console.log(`Traffic:    80% redirect, 10% page, 5% create, 5% detail`);
  console.log(`Stages:     warm(100) → ramp(300) → steady(500) → peak(800) → cool(0)`);
  console.log('');

  return { startTime: Date.now(), hasToken: !!token };
}

export function teardown(data) {
  const duration = ((Date.now() - data.startTime) / 1000).toFixed(1);
  console.log('');
  console.log(`========================================`);
  console.log(`  Mixed load test completed in ${duration}s`);
  console.log(`  Token used: ${data.hasToken ? 'yes' : 'no (redirect-only)'}`);
  console.log(`========================================`);
}
