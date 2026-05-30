import { APIError } from '@/types';

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

export const api = {
  async get<T>(endpoint: string): Promise<T> { return request<T>('GET', endpoint); },
  async post<T>(endpoint: string, body?: unknown, isFormData?: boolean): Promise<T> { return request<T>('POST', endpoint, body, isFormData); },
  async put<T>(endpoint: string, body?: unknown): Promise<T> { return request<T>('PUT', endpoint, body); },
  async delete<T>(endpoint: string): Promise<T> { return request<T>('DELETE', endpoint); },
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

async function request<T>(method: string, endpoint: string, body?: unknown, isFormData?: boolean, isRetry = false): Promise<T> {
  if (method === 'GET') {
    const key = endpoint;
    if (inflightRequests.has(key)) {
      return inflightRequests.get(key) as Promise<T>;
    }
    const promise = doRequest<T>(method, endpoint, body, isFormData, isRetry).finally(() => {
      inflightRequests.delete(key);
    });
    inflightRequests.set(key, promise);
    return promise;
  }
  return doRequest<T>(method, endpoint, body, isFormData, isRetry);
}

async function doRequest<T>(method: string, endpoint: string, body?: unknown, isFormData?: boolean, isRetry = false): Promise<T> {
  const headers: Record<string, string> = {};
  if (!isFormData && body) headers['Content-Type'] = 'application/json';

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers,
    credentials: 'include',
    body: isFormData ? (body as FormData) : body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    // Silent session refresh on 401 (skip if this IS the refresh call)
    if (res.status === 401 && !isRetry && !endpoint.includes('/auth/refresh') && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/register')) {
      if (!isRefreshing) {
        isRefreshing = true;
        const { refreshToken } = await import('./auth');
        refreshPromise = refreshToken();
      }
      const refreshed = await refreshPromise;
      isRefreshing = false;
      refreshPromise = null;

      if (refreshed) {
        return doRequest<T>(method, endpoint, body, isFormData, true);
      }

      // Refresh failed - force re-login
      const { clearAuth } = await import('./auth');
      clearAuth();
      window.location.href = '/login';
      throw new APIError('Session expired. Please log in again.', 401);
    }

    const data = await res.json().catch(() => ({}));
    throw new APIError(data.error || 'Request failed', res.status, data);
  }
  return res.json();
}
