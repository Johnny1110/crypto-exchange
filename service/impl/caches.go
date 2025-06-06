package serviceImpl

import (
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/labstack/gommon/log"
	"sync"
)

type CacheService struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewCacheService() service.ICacheService {
	return &CacheService{
		data: make(map[string]interface{}),
	}
}

func (c *CacheService) Update(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Debugf("[CacheService.Update] key=%s, value=%v", key, value)
	c.data[key] = value
}

func (c *CacheService) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	log.Debugf("[CacheService.Get] key=%s", key)
	data, exists := c.data[key]
	return data, exists
}
