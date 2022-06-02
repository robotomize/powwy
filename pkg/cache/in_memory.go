package cache

import (
	"errors"
	"sync"
	"time"
)

var ErrInvalidDuration = errors.New("invalid duration")

type Cache[T any] struct {
	data        map[string]item[T]
	expireAfter time.Duration
	mu          sync.RWMutex
}

type item[T any] struct {
	object    T
	expiresAt int64
}

func (c *item[T]) expired() bool {
	return c.expiresAt < time.Now().UnixNano()
}

func New[T any](expireAfter time.Duration) (*Cache[T], error) {
	if expireAfter < 0 {
		return nil, ErrInvalidDuration
	}

	return &Cache[T]{
		data:        make(map[string]item[T], 8),
		expireAfter: expireAfter,
	}, nil
}

func (c *Cache[T]) Lookup(name string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lookup(name)
}

func (c *Cache[T]) Set(name string, object T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[name] = item[T]{
		object:    object,
		expiresAt: time.Now().Add(c.expireAfter).UnixNano(),
	}

	return nil
}

func (c *Cache[T]) Delete(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, name)

	return nil
}

func (c *Cache[T]) lookup(name string) (T, bool) {
	var nilT T
	if item, ok := c.data[name]; ok && item.expired() {
		go c.purgeExpired(name, item.expiresAt)
		return nilT, false
	} else if ok {
		return item.object, true
	}

	return nilT, false
}

func (c *Cache[T]) purgeExpired(name string, expectedExpiryTime int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.data[name]; ok && item.expiresAt == expectedExpiryTime {
		delete(c.data, name)
	}
}
