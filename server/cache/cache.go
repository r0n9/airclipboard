package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

var (
	cache       Cache
	redisClient *redis.Client
)

const (
	CacheTypeMemory = "memory"
	CacheTypeRedis  = "redis"
)

type Config struct {
	CacheType     string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func InitCache(config Config) {
	switch config.CacheType {
	case CacheTypeRedis:
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
		cache = NewRedisCache(redisClient)
	default:
		cache = NewInMemoryCache()
	}

	go func() {
		c := cron.New()
		// 清理过期缓存，10分钟执行一次
		_, err := c.AddFunc("@every 10m", func() {
			go func() {
				cache.Clean()
				log.Printf("清理过期缓存完成，当前缓存大小：%v", cache.Size())
			}()
		})
		if err != nil {
			log.Printf("ERROR-清理过期缓存任务启动失败: %v", err)
			panic(err)
		}
		log.Println("开启定时任务，10分钟执行一次，清理过期缓存")
		c.Start()
	}()
}

type Cache interface {
	Get(key string) ([]*Message, bool)
	Set(key string, data []*Message, duration time.Duration)
	Delete(key string)
	GetAllKeys() []string
	Size() int
	Clean()
	GetExpireAt(key string) string

	SetIp2BoardName(ip, boardName string, duration time.Duration)
	GetIp2BoardName(ip string) (string, bool)
}

type Message struct {
	Id       string `json:"id"`
	Content  string `json:"content"`
	Time     string `json:"time"`
	Ip       string `json:"ip"`
	IsFile   bool   `json:"isFile"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

type cachedItem struct {
	Data       []*Message
	Expiration int64
}
type cachedBoardName struct {
	BoardName  string
	Expiration int64
}

func GetFromCache(key string) ([]*Message, bool) {
	return cache.Get(key)
}

func SetToCache(key string, data []*Message, duration time.Duration) {
	cache.Set(key, data, duration)
}

func DeleteFromCache(key string) {
	cache.Delete(key)
}

func GetExpireAt(key string) string {
	return cache.GetExpireAt(key)
}

func GetAllKeys() []string {
	return cache.GetAllKeys()
}

func CacheSize() int {
	return cache.Size()
}

func GetBoardNameFromCache(ip string) (string, bool) {
	return cache.GetIp2BoardName(ip)
}

func SetBoardNameToCache(ip, boardName string, duration time.Duration) {
	cache.SetIp2BoardName(ip, boardName, duration)
}
