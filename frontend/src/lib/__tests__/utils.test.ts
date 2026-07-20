import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { cn, timeAgo, formatCurrency, escapeHtml } from '../utils';

describe('cn', () => {
  it('merges class names', () => {
    const result = cn('foo', 'bar');
    expect(result).toBe('foo bar');
  });

  it('handles conditional classes', () => {
    const result = cn('foo', false && 'bar', 'baz');
    expect(result).toBe('foo baz');
  });

  it('deduplicates tailwind classes', () => {
    const result = cn('p-2', 'p-4');
    expect(result).toBe('p-4');
  });
});

describe('timeAgo', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2025-06-15T12:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "Just now" for empty string', () => {
    expect(timeAgo('')).toBe('Just now');
  });

  it('returns "Just now" for recent dates', () => {
    expect(timeAgo('2025-06-15T11:59:30Z')).toBe('Just now');
  });

  it('returns minutes ago', () => {
    expect(timeAgo('2025-06-15T11:50:00Z')).toBe('10m ago');
  });

  it('returns hours ago', () => {
    expect(timeAgo('2025-06-15T09:00:00Z')).toBe('3h ago');
  });

  it('returns days ago', () => {
    expect(timeAgo('2025-06-13T12:00:00Z')).toBe('2d ago');
  });
});

describe('formatCurrency', () => {
  it('formats amount in Naira', () => {
    expect(formatCurrency(1000)).toBe('₦1,000');
  });

  it('formats zero', () => {
    expect(formatCurrency(0)).toBe('₦0');
  });

  it('formats large amounts', () => {
    expect(formatCurrency(1000000)).toBe('₦1,000,000');
  });
});

describe('escapeHtml', () => {
  it('escapes HTML entities', () => {
    expect(escapeHtml('<script>alert("xss")</script>')).toBe(
      '&lt;script&gt;alert("xss")&lt;/script&gt;'
    );
  });

  it('returns plain text unchanged', () => {
    expect(escapeHtml('hello world')).toBe('hello world');
  });
});
