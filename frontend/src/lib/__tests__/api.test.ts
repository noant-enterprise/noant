import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { api } from '../api';
import { APIError } from '@/types';

const mockFetch = vi.fn();

beforeEach(() => {
  vi.stubGlobal('fetch', mockFetch);
  mockFetch.mockReset();
});

afterEach(() => {
  vi.restoreAllMocks();
});

function jsonResponse(data: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(data),
    headers: new Headers(),
  };
}

describe('api.get', () => {
  it('sends GET request to correct endpoint', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ id: 1 }));

    const result = await api.get<{ id: number }>('/users/1');

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/users/1'),
      expect.objectContaining({ method: 'GET' })
    );
    expect(result).toEqual({ id: 1 });
  });

  it('throws APIError on failure', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Not found' }, 404)
    );

    await expect(api.get('/missing')).rejects.toThrow(APIError);
  });

  it('deduplicates concurrent GET requests', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ id: 1 }));

    const [a, b] = await Promise.all([api.get('/users/1'), api.get('/users/1')]);

    expect(mockFetch).toHaveBeenCalledTimes(1);
    expect(a).toEqual(b);
  });
});

describe('api.post', () => {
  it('sends POST with JSON body', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ ok: true }, 201));

    await api.post('/items', { name: 'test' });

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/items'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'test' }),
        credentials: 'include',
      })
    );
  });

  it('sends FormData without Content-Type header', async () => {
    const formData = new FormData();
    formData.append('file', new Blob(), 'test.png');
    mockFetch.mockResolvedValue(jsonResponse({ url: 'ok' }));

    await api.post('/upload', formData, true);

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/upload'),
      expect.objectContaining({
        method: 'POST',
        body: formData,
      })
    );
    const callHeaders = mockFetch.mock.calls[0]![1]!.headers;
    expect(callHeaders['Content-Type']).toBeUndefined();
  });
});

describe('api.put', () => {
  it('sends PUT request with body', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ updated: true }));

    const result = await api.put('/items/1', { name: 'updated' });

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/items/1'),
      expect.objectContaining({ method: 'PUT' })
    );
    expect(result).toEqual({ updated: true });
  });
});

describe('api.delete', () => {
  it('sends DELETE request', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ deleted: true }));

    const result = await api.delete('/items/1');

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/items/1'),
      expect.objectContaining({ method: 'DELETE' })
    );
    expect(result).toEqual({ deleted: true });
  });
});

describe('retry on 429', () => {
  it('retries up to 3 times on rate limit', async () => {
    const rateLimited = {
      ok: false,
      status: 429,
      json: () => Promise.resolve({ error: 'rate limited' }),
      headers: new Headers({ 'Retry-After': '0' }),
    };

    mockFetch
      .mockResolvedValueOnce(rateLimited)
      .mockResolvedValueOnce(rateLimited)
      .mockResolvedValueOnce(rateLimited)
      .mockResolvedValueOnce(jsonResponse({ ok: true }));

    const result = await api.get<{ ok: boolean }>('/slow-endpoint');

    expect(mockFetch).toHaveBeenCalledTimes(4);
    expect(result).toEqual({ ok: true });
  });
});
