package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"noant/config"
)

type CacheEntry struct {
	Value     string
	ExpiresAt time.Time
	CreatedAt time.Time
	HitCount  int64
}

type Cache struct {
	mu         sync.RWMutex
	store      map[string]*CacheEntry
	maxKeys    int
	defaultTTL time.Duration
	redis      *RedisClient
	stats      CacheStats
}

type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	CurrentSize int64
}

func NewCache(cfg *config.Config, redis *RedisClient) *Cache {
	c := &Cache{
		store:      make(map[string]*CacheEntry),
		maxKeys:    cfg.CacheMaxKeys,
		defaultTTL: cfg.CacheTTL,
	}
	if redis != nil {
		c.redis = redis
	}
	return c
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	c.mu.RLock()
	entry, found := c.store[key]
	c.mu.RUnlock()
	if found {
		if time.Now().Before(entry.ExpiresAt) {
			c.mu.Lock()
			c.stats.Hits++
			c.mu.Unlock()
			return entry.Value, true
		}
		c.mu.Lock()
		delete(c.store, key)
		c.stats.Evictions++
		c.mu.Unlock()
	}
	if c.redis != nil {
		val, err := c.redis.Get(ctx, "cache:"+key)
		if err == nil && val != "" {
			c.mu.Lock()
			c.store[key] = &CacheEntry{Value: val, ExpiresAt: time.Now().Add(c.defaultTTL), CreatedAt: time.Now()}
			c.stats.Hits++
			c.mu.Unlock()
			return val, true
		}
	}
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
	return "", false
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	c.mu.Lock()
	if len(c.store) >= c.maxKeys {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.store {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}
		if oldestKey != "" {
			delete(c.store, oldestKey)
			c.stats.Evictions++
		}
	}
	c.store[key] = &CacheEntry{Value: value, ExpiresAt: time.Now().Add(ttl), CreatedAt: time.Now()}
	c.stats.CurrentSize = int64(len(c.store))
	c.mu.Unlock()
	if c.redis != nil {
		_ = c.redis.Set(context.Background(), "cache:"+key, value, ttl)
	}
}

func (c *Cache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
	if c.redis != nil {
		_ = c.redis.Delete(context.Background(), "cache:"+key)
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	c.store = make(map[string]*CacheEntry)
	c.stats = CacheStats{}
	c.mu.Unlock()
}

func (c *Cache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	if val, found := c.Get(ctx, key); found {
		return val, nil
	}
	val, err := fn()
	if err != nil {
		return "", err
	}
	c.Set(ctx, key, val, ttl)
	return val, nil
}

func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func GetCacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, p := range parts {
		key = fmt.Sprintf("%s:%s", key, p)
	}
	return key
}