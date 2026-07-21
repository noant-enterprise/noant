import http from 'k6/http';
import { check, sleep, group, randomIntBetween } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 5 },
    { duration: '10s', target: 200 },
    { duration: '30s', target: 200 },
    { duration: '10s', target: 5 },
    { duration: '2m', target: 5 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],
    http_req_failed: ['rate<0.15'],
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export function setup() {
  const accounts = [];
  for (let i = 0; i < 200; i++) {
    const email = `spike_vu${i}_${Date.now()}@noant.test`;
    const password = 'SpikeTest123!';

    http.post(`${BASE_URL}/auth/register`, JSON.stringify({
      email,
      password,
      name: `Spike User ${i}`,
    }), {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'setup/register' },
    });

    const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
      email,
      password,
    }), {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'setup/login' },
    });

    const token = loginRes.json('data.token') || loginRes.json('token');
    if (token) {
      accounts.push({ token, index: i });
    }
  }
  return { accounts };
}

export default function (data) {
  const account = data.accounts[__VU % data.accounts.length];
  if (!account) return;

  const headers = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${account.token}`,
  };

  group('Health', () => {
    const res = http.get(`${BASE_URL}/health`, {
      tags: { name: '/health' },
    });
    check(res, {
      'health status 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  const roll = Math.random();

  if (roll < 0.6) {
    group('Spike: conversations', () => {
      const res = http.get(`${BASE_URL}/chats/conversations`, {
        headers,
        tags: { name: '/chats/conversations' },
      });
      check(res, {
        'conversations status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 2));

    group('Spike: analytics', () => {
      const res = http.get(`${BASE_URL}/analytics/overview`, {
        headers,
        tags: { name: '/analytics/overview' },
      });
      check(res, {
        'analytics status 200': (r) => r.status === 200,
      });
    });
  } else if (roll < 0.85) {
    group('Spike: direct chat', () => {
      const res = http.post(`${BASE_URL}/chats/direct-chat`, JSON.stringify({
        message: `Spike test message from VU ${__VU}`,
      }), {
        headers,
        tags: { name: '/chats/direct-chat' },
      });
      check(res, {
        'direct-chat status 200': (r) => r.status === 200,
      });
    });
  } else {
    group('Spike: auth/me', () => {
      const res = http.get(`${BASE_URL}/auth/me`, {
        headers,
        tags: { name: '/auth/me' },
      });
      check(res, {
        'auth/me status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 2));

    group('Spike: settings', () => {
      const res = http.get(`${BASE_URL}/settings/profile`, {
        headers,
        tags: { name: '/settings/profile' },
      });
      check(res, {
        'settings status 200': (r) => r.status === 200,
      });
    });
  }

  sleep(randomIntBetween(1, 3));
}

export function teardown(data) {
  for (const account of data.accounts) {
    http.del(`${BASE_URL}/auth/me`, null, {
      headers: { Authorization: `Bearer ${account.token}` },
      tags: { name: 'cleanup' },
    });
  }
}
