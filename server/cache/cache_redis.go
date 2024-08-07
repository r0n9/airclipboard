package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var prefixBoard = "sync-board."
var prefixIp = "ip:"

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {

	statusCmd := client.Ping(context.Background())
	if statusCmd.Err() != nil {
		log.Fatalf("ERROR: Redis ping failed: %v", statusCmd.Err())
		return nil
	}

	return &RedisCache{client: client}
}

func (c *RedisCache) Get(key string) ([]*Message, bool) {
	val, err := c.client.Get(context.Background(), keyPrefix(prefixBoard, key)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	} else if err != nil {
		log.Printf("ERROR: Redis GET failed: %v", err)
		return nil, false
	}

	var data []*Message
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		log.Printf("ERROR: Redis unmarshal failed: %v", err)
		return nil, false
	}

	return data, true
}

func (c *RedisCache) GetExpireAt(key string) string {
	ctx := context.Background()
	ttl, err := c.client.TTL(ctx, keyPrefix(prefixBoard, key)).Result()
	if err != nil || ttl == -2 { // -2 means the key does not exist
		return ""
	}

	if ttl == -1 { // -1 means the key has no expiration
		return ""
	}

	expireAt := time.Now().Add(ttl)
	return expireAt.Format("2006-01-02 15:04:05")
}

func (c *RedisCache) Set(key string, data []*Message, duration time.Duration) {
	val, err := json.Marshal(data)
	if err != nil {
		log.Printf("ERROR: Redis marshal failed: %v", err)
		return
	}

	err = c.client.Set(context.Background(), keyPrefix(prefixBoard, key), val, duration).Err()
	if err != nil {
		log.Printf("ERROR: Redis SET failed: %v", err)
	}
}

func (c *RedisCache) Delete(key string) {
	err := c.client.Del(context.Background(), keyPrefix(prefixBoard, key)).Err()
	if err != nil {
		log.Printf("ERROR: Redis DEL failed: %v", err)
	}
}

func (c *RedisCache) GetAllKeys() []string {
	var cursor uint64
	var keys []string

	for {
		var k []string
		var err error
		k, cursor, err = c.client.Scan(context.Background(), cursor, keyPrefix(prefixBoard, "*"), 1000).Result()
		if err != nil {
			log.Printf("ERROR: Redis SCAN failed: %v", err)
			break
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}

	removedPrefixKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		removedPrefixKeys = append(removedPrefixKeys, removeKeyPrefix(prefixBoard, k))
	}

	return removedPrefixKeys
}

func (c *RedisCache) Size() int {
	keys := c.GetAllKeys()
	return len(keys)
}

func (c *RedisCache) Clean() {
	return
}

func keyPrefix(prefix, key string) string {
	return fmt.Sprintf("%s%s", prefix, key)
}

func removeKeyPrefix(prefix, key string) string {
	return key[len(prefix):]
}

func (c *RedisCache) SetIp2BoardName(ip, boardName string, duration time.Duration) {
	err := c.client.Set(context.Background(), keyPrefix(prefixIp, ip), boardName, duration).Err()
	if err != nil {
		log.Printf("ERROR: Redis SET failed: %v", err)
	}
}

func (c *RedisCache) GetIp2BoardName(ip string) (string, bool) {
	val, err := c.client.Get(context.Background(), keyPrefix(prefixIp, ip)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false
	} else if err != nil {
		log.Printf("ERROR: Redis GET failed: %v", err)
		return "", false
	}
	return val, true
}
