package cache

import (
	"log"
	"sync"
	"time"
)

type InMemoryCache struct {
	cache          map[string]cachedItem
	cacheBoardName map[string]cachedBoardName
	lock           sync.Mutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache:          make(map[string]cachedItem),
		cacheBoardName: make(map[string]cachedBoardName),
	}
}

func (c *InMemoryCache) Get(key string) ([]*Message, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, found := c.cache[key]
	if !found {
		return nil, false
	}

	if item.Expiration < time.Now().UnixNano() {
		delete(c.cache, key)
		return nil, false
	}

	return item.Data, true
}

func (c *InMemoryCache) GetExpireAt(key string) string {
	c.lock.Lock()
	defer c.lock.Unlock()

	item, found := c.cache[key]
	if !found {
		return ""
	}

	expireAt := time.Unix(0, item.Expiration)
	if expireAt.Before(time.Now()) {
		delete(c.cache, key)
		return ""
	}

	return expireAt.Format("2006-01-02 15:04:05")
}

func (c *InMemoryCache) Set(key string, data []*Message, duration time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.cache[key] = cachedItem{
		Data:       data,
		Expiration: expiration,
	}
}

func (c *InMemoryCache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.cache, key)
}

func (c *InMemoryCache) GetAllKeys() []string {
	c.lock.Lock()
	defer c.lock.Unlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}

func (c *InMemoryCache) Size() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return len(c.cache)
}
func (c *InMemoryCache) Clean() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k, v := range c.cache {
		// 清理缓存中那些已经过期的项
		if v.Expiration < time.Now().UnixNano() {
			delete(c.cache, k)
			continue
		}
	}

	for k, v := range c.cacheBoardName {
		// 清理缓存中那些已经过期的项
		if v.Expiration < time.Now().UnixNano() {
			delete(c.cacheBoardName, k)
			continue
		}
	}

	log.Printf("清理过期缓存完成，当前缓存大小：%v", c.Size())
}

func (c *InMemoryCache) SetIp2BoardName(ip, boardName string, duration time.Duration) {
	expiration := time.Now().Add(duration).UnixNano()
	c.cacheBoardName[ip] = cachedBoardName{
		BoardName:  boardName,
		Expiration: expiration,
	}
}

func (c *InMemoryCache) GetIp2BoardName(ip string) (string, bool) {
	item, found := c.cacheBoardName[ip]
	if !found {
		return "", false
	}

	if item.Expiration < time.Now().UnixNano() {
		delete(c.cacheBoardName, ip)
		return "", false
	}

	return item.BoardName, true
}
