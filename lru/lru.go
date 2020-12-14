package lru

import (
	"container/list"
)

// Cache is an LRU cache
type Cache struct {
	// MaxEntries is the maximum number of cache concurrent access.
	MaxEntries int64

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache
	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
}

// Key may be any value that is comparable.
type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

// New creates a new Cache
// If the maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller
func New(maxEntries int64) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	// if exist move to front
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		ele.Value.(*entry).value = value
		return
	}

	// not exist add to front
	ele := c.ll.PushFront(&entry{
		key:   key,
		value: value,
	})
	c.cache[key] = ele
	if c.MaxEntries != 0 && int64(c.ll.Len()) > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ee, ok := c.cache[key]; ok {
		// move to front
		c.ll.MoveToFront(ee)
		return ee.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}

	if ee, ok := c.cache[key]; ok {
		c.removeElement(ee)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}

	ee := c.ll.Back()
	if ee != nil {
		c.removeElement(ee)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}
