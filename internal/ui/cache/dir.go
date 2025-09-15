package cache

import (
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/panels"
)

type DirCache struct {
	m     map[string][]panels.Entry
	order []string
	max   int
}

func NewDirCache(max int) DirCache {
	if max <= 0 {
		max = 64
	}
	return DirCache{m: make(map[string][]panels.Entry), max: max}
}

func (c *DirCache) Get(key string) []panels.Entry { return c.m[key] }
func (c *DirCache) Has(key string) bool           { _, ok := c.m[key]; return ok }

func (c *DirCache) Put(key string, val []panels.Entry) {
	if val == nil {
		return
	}
	if _, ok := c.m[key]; ok {
		return
	}
	if len(c.order) >= c.max {
		old := c.order[0]
		delete(c.m, old)
		c.order = c.order[1:]
	}
	c.order = append(c.order, key)
	c.m[key] = val
}
