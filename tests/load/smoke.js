import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  vus: 1,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.1'],
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';
const TEST_USER = {
  email: `smoke_test_${Date.now()}@noant.test`,
  password: 'SmokeTest123!',
  name: 'Smoke Test User',
};

export function setup() {
  const regRes = http.post(`${BASE_URL}/auth/register`, JSON.stringify(TEST_USER), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: '/auth/register' },
  });

  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: TEST_USER.email,
    password: TEST_USER.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: '/auth/login' },
  });

  const token = loginRes.json('data.token') || loginRes.json('token');
  if (!token) {
    throw new Error(`Login failed: ${loginRes.status} ${loginRes.body}`);
  }

  return { token };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.token}`,
  };

  group('Auth', () => {
    const res = http.get(`${BASE_URL}/auth/me`, {
      headers,
      tags: { name: '/auth/me' },
    });
    check(res, {
      '/auth/me status 200': (r) => r.status === 200,
      '/auth/me has user': (r) => r.json('data') !== undefined,
    });
  });

  sleep(1);

  group('Conversations', () => {
    const res = http.get(`${BASE_URL}/chats/conversations`, {
      headers,
      tags: { name: '/chats/conversations' },
    });
    check(res, {
      '/chats/conversations status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  group('Direct Chat', () => {
    const res = http.post(`${BASE_URL}/chats/direct-chat`, JSON.stringify({
      message: 'Hello, testing',
    }), {
      headers,
      tags: { name: '/chats/direct-chat' },
    });
    check(res, {
      '/chats/direct-chat status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  group('Training', () => {
    const res = http.get(`${BASE_URL}/training/categories`, {
      headers,
      tags: { name: '/training/categories' },
    });
    check(res, {
      '/training/categories status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  group('Analytics', () => {
    const res = http.get(`${BASE_URL}/analytics/overview`, {
      headers,
      tags: { name: '/analytics/overview' },
    });
    check(res, {
      '/analytics/overview status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  group('Settings', () => {
    const res = http.get(`${BASE_URL}/settings/profile`, {
      headers,
      tags: { name: '/settings/profile' },
    });
    check(res, {
      '/settings/profile status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  group('Health', () => {
    const res = http.get(`${BASE_URL}/health`, {
      tags: { name: '/health' },
    });
    check(res, {
      '/health status 200': (r) => r.status === 200,
    });
  });
}

export function teardown(data) {
  http.del(`${BASE_URL}/auth/me`, null, {
    headers: { Authorization: `Bearer ${data.token}` },
    tags: { name: 'cleanup' },
  });
}
