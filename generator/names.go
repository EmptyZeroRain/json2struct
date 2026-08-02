package generator

import "sync"

type nameCache struct {
	mu     sync.RWMutex
	values map[string]string
}
type NameCache = nameCache

func NewNameCache() *NameCache { return newNameCache() }

func newNameCache() *nameCache { return &nameCache{values: make(map[string]string)} }
func (c *nameCache) get(key string, makeName func(string) string) string {
	c.mu.RLock()
	v, ok := c.values[key]
	c.mu.RUnlock()
	if ok {
		return v
	}
	v = makeName(key)
	c.mu.Lock()
	if old, ok := c.values[key]; ok {
		v = old
	} else {
		c.values[key] = v
	}
	c.mu.Unlock()
	return v
}
