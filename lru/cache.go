package lru

import (
	"container/list"
	"sync"
)

type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	ll       *list.List
	items    map[K]*list.Element
	capacity int
}
type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}
	return &Cache[K, V]{
		ll:       list.New(),
		items:    make(map[K]*list.Element),
		capacity: capacity,
	}
}
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		el.Value = entry[K, V]{key: key, value: value}
		c.ll.MoveToFront(el)
		return
	}

	el := c.ll.PushFront(entry[K, V]{key: key, value: value})
	c.items[key] = el

	if c.ll.Len() > c.capacity {
		last := c.ll.Back()
		if last != nil {
			kv := last.Value.(entry[K, V])
			delete(c.items, kv.key)
			c.ll.Remove(last)
		}
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		kv := el.Value.(entry[K, V])
		return kv.value, true
	}
	var zero V
	return zero, false
}

func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll = list.New()
	c.items = make(map[K]*list.Element, c.capacity)
}

func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *Cache[K, V]) Cap() int {
	return c.capacity
}
