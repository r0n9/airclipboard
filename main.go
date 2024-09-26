package main

import (
	"airclipboard/common"
	"airclipboard/server"
	"airclipboard/server/cache"
	"airclipboard/slog"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"
	"unicode"
)

//go:embed templates
var content embed.FS

func main() {
	cacheType := flag.String("cache-type", "memory", "Cache type (memory or redis)")
	redisAddr := flag.String("redis-addr", "localhost:6379", "Address of the Redis server")
	redisPassword := flag.String("redis-password", "******", "Password for the Redis server")
	redisDB := flag.Int("redis-db", 0, "Redis database number")

	flag.Parse()

	slog.Init() // 日志初始化

	config := cache.Config{
		CacheType:     *cacheType,
		RedisAddr:     *redisAddr,
		RedisPassword: *redisPassword,
		RedisDB:       *redisDB,
	}
	cache.InitCache(config)

	r := gin.New()
	initRoute(r)
	base := fmt.Sprintf("%s:%d", "0.0.0.0", 18128)
	log.Printf("Start server @ %s", base)
	srv := &http.Server{Addr: base, Handler: r}
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start: %s", err.Error())
	}
}

func initRoute(e *gin.Engine) {
	Cors(e)
	peerServer := server.NewPeerServer()
	e.GET("/server/webrtc", func(c *gin.Context) {
		peerServer.HandleConnection(c)
	})

	// 静态文件路由
	e.GET("/service-worker.js", func(c *gin.Context) {
		data, err := content.ReadFile("templates/service-worker.js")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/javascript", data)
	})

	tmpl := template.Must(template.New("").ParseFS(content, "templates/*.html"))
	e.SetHTMLTemplate(tmpl)

	folders := []string{"css", "images", "scripts", "sounds"}

	for i, folder := range folders {
		folder = "templates/" + folder
		sub, err := fs.Sub(content, folder)
		if err != nil {
			log.Fatalf("can't find folder: %s", folder)
		}
		e.StaticFS(fmt.Sprintf("/%s/", folders[i]), http.FS(sub))
	}

	e.GET("/", func(c *gin.Context) {
		realIp := server.LogApiRequestIP(c, "Index", -1)

		var board string

		if cookie, err := c.Request.Cookie("board"); err == nil && cookie.Value != "" {
			board = cookie.Value
		} else {
			// 默认同网络内的同名板块
			if boardExist, ok := cache.GetBoardNameFromCache(realIp); ok {
				board = boardExist
			} else {
				// 生成6位随机字符串，只包含数字和小写字母
				board = common.RandString(6)
			}
			c.Header("Set-Cookie", "board="+board+";SameSite=Strict;Secure")
		}
		cache.SetBoardNameToCache(realIp, board, time.Hour*48)
		c.HTML(200, "index.html", gin.H{"Board": board})
	})

	e.GET("/:board", func(c *gin.Context) {
		board := c.Param("board")
		board = truncateBoard(board)
		if board == "index." {
			// 生成6位随机字符串，只包含数字和小写字母
			board = common.RandString(6)
		}
		c.Header("Set-Cookie", "board="+board+";SameSite=Strict;Secure")
		c.HTML(200, "index.html", gin.H{"Board": board})
	})

	mfApi := e.Group("/boardapi")
	mfApi.GET("/:board", server.FetchBoard)
	mfApi.POST("/:board", server.AddMessage)
	mfApi.DELETE("/:board/:id", server.DeleteMessage)
	mfApi.GET("/:board/:id", server.GetMessage)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"*"}
	config.AllowMethods = []string{"*"}
	r.Use(cors.New(config))
}

func truncateBoard(board string) string {
	runes := []rune(board)

	// Initialize counters
	count := 0
	chineseCount := 0

	for i, r := range runes {
		if isChinese(r) {
			chineseCount++
		}

		count++
		if count == 6 || chineseCount == 4 {
			return string(runes[:i+1])
		}
	}

	return string(runes)
}

func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
