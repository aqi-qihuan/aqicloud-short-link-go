/**
 * k6 冒烟测试 — 快速验证服务可用性
 *
 * 用途: 压测前的预检，确认各接口返回正常
 * 耗时: ~10 秒
 *
 * 使用方法:
 *   k6 run test/stress/k6_smoke.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';

export const options = {
  vus: 1,
  duration: '10s',
  thresholds: {
    'http_req_failed': ['rate<0.5'],  // 允许一半失败 (测试数据可能不存在)
  },
};

export default function () {
  // 测试 1: Gateway 可达
  const r1 = http.get(`${BASE_URL}/`, { redirects: 0, timeout: '5s' });
  check(r1, { 'gateway reachable': (r) => r.status !== 0 });

  // 测试 2: 短链重定向 (期望 302 或 404)
  const r2 = http.get(`${BASE_URL}/test00`, { redirects: 0, timeout: '5s' });
  check(r2, {
    'short link endpoint ok': (r) => [302, 404, 400, 403].includes(r.status),
    'response < 100ms': (r) => r.timings.duration < 100,
  });

  // 测试 3: Gateway 限流 header
  const r3 = http.get(`${BASE_URL}/test01`, { redirects: 0, timeout: '5s' });
  check(r3, { 'no 429 at low rate': (r) => r.status !== 429 });

  // 测试 4: 路由到后端服务
  const r4 = http.get(`${BASE_URL}/link-server/api/domain/v1/list`, { timeout: '5s' });
  check(r4, { 'backend proxy works': (r) => r.status !== 0 });

  sleep(0.5);
}

export function setup() {
  console.log(`Smoke test target: ${BASE_URL}`);
}
