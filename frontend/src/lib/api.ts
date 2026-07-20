import { APIError } from '@/types';

let API_BASE = import.meta.env.VITE_API_URL || '/api/v1';
if (API_BASE && !API_BASE.startsWith('/') && !API_BASE.startsWith('http://') && !API_BASE.startsWith('https://')) {
  API_BASE = `https://${API_BASE}`;
}
if (API_BASE.startsWith('http') && !API_BASE.endsWith('/api/v1')) {
  API_BASE = `${API_BASE.replace(/\/$/, '')}/api/v1`;
}

export const api = {
  async get<T>(endpoint: string): Promise<T> { return request<T>('GET', endpoint); },
  async post<T>(endpoint: string, body?: unknown, isFormData?: boolean): Promise<T> { return request<T>('POST', endpoint, body, isFormData); },
  async put<T>(endpoint: string, body?: unknown): Promise<T> { return request<T>('PUT', endpoint, body); },
  async delete<T>(endpoint: string): Promise<T> { return request<T>('DELETE', endpoint); },
  streamPost(endpoint: string, body: unknown, onChunk: (chunk: string) => void, onDone?: (meta?: Record<string, unknown>) => void, onError?: (err: Error) => void): AbortController {
    const controller = new AbortController();
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };

    fetch(`${API_BASE}${endpoint}`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: JSON.stringify(body),
      signal: controller.signal,
    }).then(async (res) => {
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        onError?.(new APIError(data.error || 'Stream request failed', res.status));
        return;
      }
      const reader = res.body?.getReader();
      if (!reader) { onError?.(new Error('No response body')); return; }
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue;
          const data = line.slice(6);
          if (data === '[DONE]') {
            onDone?.();
            return;
          }
          if (data === '[DONE]') { onDone?.(); return; }
          if (data.startsWith('[DONE]')) {
            const metaStr = data.slice(7);
            try { onDone?.(JSON.parse(metaStr)); } catch { onDone?.(); }
            return;
          }
          if (data === '[ERROR]') {
            onError?.(new Error('Server error during streaming'));
            return;
          }
          onChunk(data);
        }
      }
      onDone?.();
    }).catch((err) => {
      if (err.name !== 'AbortError') {
        onError?.(err);
      }
    });

    return controller;
  },
};

export interface APIClient {
  get: typeof api.get;
  post: typeof api.post;
  put: typeof api.put;
  delete: typeof api.delete;
}

let isRefreshing = false;
let refreshPromise: Promise<string | null> | null = null;
const inflightRequests = new Map<string, Promise<any>>();

async function request<T>(method: string, endpoint: string, body?: unknown, isFormData?: boolean, retryCount = 0): Promise<T> {
  if (method === 'GET') {
    const key = endpoint;
    if (inflightRequests.has(key)) {
      return inflightRequests.get(key) as Promise<T>;
    }
    const promise = doRequest<T>(method, endpoint, body, isFormData, retryCount).finally(() => {
      inflightRequests.delete(key);
    });
    inflightRequests.set(key, promise);
    return promise;
  }
  return doRequest<T>(method, endpoint, body, isFormData, retryCount);
}

async function doRequest<T>(method: string, endpoint: string, body?: unknown, isFormData?: boolean, retryCount = 0): Promise<T> {
  const headers: Record<string, string> = {};
  if (!isFormData && body) headers['Content-Type'] = 'application/json';

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers,
    credentials: 'include',
    body: isFormData ? (body as FormData) : body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    // Retry on 429 rate limit with exponential backoff (up to 3 attempts)
    if (res.status === 429 && retryCount < 3) {
      const retryAfterHeader = res.headers.get('Retry-After');
      const delay = retryAfterHeader
        ? parseInt(retryAfterHeader, 10) * 1000
        : Math.pow(2, retryCount) * 1000; // 1s, 2s, 4s
      await new Promise(resolve => setTimeout(resolve, delay));
      return doRequest<T>(method, endpoint, body, isFormData, retryCount + 1);
    }

    // Silent session refresh on 401 (skip if this IS the refresh call)
    if (res.status === 401 && retryCount === 0 && !endpoint.includes('/auth/refresh') && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/register')) {
      if (!isRefreshing) {
        isRefreshing = true;
        const { refreshToken } = await import('./auth');
        refreshPromise = refreshToken();
      }
      const refreshed = await refreshPromise;
      isRefreshing = false;
      refreshPromise = null;

      if (refreshed) {
        return doRequest<T>(method, endpoint, body, isFormData, retryCount + 1);
      }

      // Refresh failed - force re-login
      const { clearAuth } = await import('./auth');
      clearAuth();

      const cleanPath = window.location.pathname.replace(/\/$/, '') || '/';
      const publicPaths = ['/', '/login', '/signup', '/forgot-password', '/reset-password'];
      if (!publicPaths.includes(cleanPath)) {
        window.location.href = '/login';
      }
      throw new APIError('Session expired. Please log in again.', 401);
    }

    const data = await res.json().catch(() => ({}));
    throw new APIError(data.error || 'Request failed', res.status, data);
  }
  return res.json();
}
