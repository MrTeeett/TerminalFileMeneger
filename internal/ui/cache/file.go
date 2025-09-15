package cache

import (
    "github.com/MrTeeett/TerminalFileMeneger/internal/ui/preview"
)

type FileCache struct {
	m     map[string]preview.Result
	order []string
	max   int
}

func NewFileCache(max int) FileCache {
	if max <= 0 {
		max = 64
	}
	return FileCache{m: make(map[string]preview.Result), max: max}
}

func (c *FileCache) Get(key string) (preview.Result, bool) { v, ok := c.m[key]; return v, ok }
func (c *FileCache) Has(key string) bool                   { _, ok := c.m[key]; return ok }

func (c *FileCache) Put(key string, val preview.Result) {
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
