import http from 'k6/http';
import { check, sleep, group, randomIntBetween } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],
    http_req_failed: ['rate<0.05'],
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export function setup() {
  const accounts = [];
  for (let i = 0; i < 50; i++) {
    const email = `stress_vu${i}_${Date.now()}@noant.test`;
    const password = 'StressTest123!';

    http.post(`${BASE_URL}/auth/register`, JSON.stringify({
      email,
      password,
      name: `Stress User ${i}`,
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

  const roll = Math.random();

  if (roll < 0.7) {
    group('Browse: conversations', () => {
      const res = http.get(`${BASE_URL}/chats/conversations`, {
        headers,
        tags: { name: '/chats/conversations' },
      });
      check(res, {
        'conversations status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 3));

    group('Browse: analytics', () => {
      const res = http.get(`${BASE_URL}/analytics/overview`, {
        headers,
        tags: { name: '/analytics/overview' },
      });
      check(res, {
        'analytics status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 3));

    group('Browse: training', () => {
      const res = http.get(`${BASE_URL}/training/categories`, {
        headers,
        tags: { name: '/training/categories' },
      });
      check(res, {
        'training status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 2));
  } else if (roll < 0.9) {
    group('Chat: send message', () => {
      const msgRes = http.post(`${BASE_URL}/chats/direct-chat`, JSON.stringify({
        message: `Stress test message from VU ${__VU}`,
      }), {
        headers,
        tags: { name: '/chats/direct-chat' },
      });
      check(msgRes, {
        'direct-chat status 200': (r) => r.status === 200,
      });

      if (msgRes.status === 200) {
        const conversationId = msgRes.json('data.conversationId') || msgRes.json('data.id');

        sleep(randomIntBetween(2, 5));

        if (conversationId) {
          const historyRes = http.get(`${BASE_URL}/chats/conversations/${conversationId}/messages`, {
            headers,
            tags: { name: '/chats/conversations/:id/messages' },
          });
          check(historyRes, {
            'conversation messages status 200': (r) => r.status === 200,
          });
        }
      }
    });

    sleep(randomIntBetween(1, 3));
  } else {
    group('Admin: settings', () => {
      const res = http.get(`${BASE_URL}/settings/profile`, {
        headers,
        tags: { name: '/settings/profile' },
      });
      check(res, {
        'settings status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 3));

    group('Admin: credits/balance', () => {
      const res = http.get(`${BASE_URL}/credits/balance`, {
        headers,
        tags: { name: '/credits/balance' },
      });
      check(res, {
        'credits balance status 200': (r) => r.status === 200,
      });
    });

    sleep(randomIntBetween(1, 2));
  }
}

export function teardown(data) {
  for (const account of data.accounts) {
    http.del(`${BASE_URL}/auth/me`, null, {
      headers: { Authorization: `Bearer ${account.token}` },
      tags: { name: 'cleanup' },
    });
  }
}
