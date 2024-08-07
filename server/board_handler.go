package server

import (
	"airclipboard/common"
	"airclipboard/server/cache"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"time"
)

const (
	MaxMessageSize = 5
	MaxBoardSize   = 30
)

type MessageReq struct {
	Content string `json:"content"`
}

type BoardInfo struct {
	Board    string           `json:"board"`
	ExpireAt string           `json:"expireAt"`
	Messages []*cache.Message `json:"messages"`
}

func LogApiRequestIP(c *gin.Context, apiName string, userId int64) string {
	realIP := c.GetHeader("CF-Connecting-IP")
	// 如果 CF-Connecting-IP 不存在，尝试获取 X-Forwarded-For 头部
	if realIP == "" {
		realIP = c.GetHeader("X-Forwarded-For")
	}

	if realIP == "" {
		realIP = c.ClientIP() // 如果 Cloudflare 的头部不存在，则使用 Gin 上下文提供的方法获取 IP 地址
	}

	log.Printf("Request %v from IP: %s", apiName, realIP)

	return realIP
}

func AddMessage(c *gin.Context) {
	board := c.Param("board")
	if board == "" {
		common.ErrorStrResp(c, "board not found ！", http.StatusNotFound)
		return
	}

	realIp := LogApiRequestIP(c, "AddMessage: "+board, -1)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("添加标签页失败，err=%v", err)
		common.ErrorStrResp(c, "请求失败！", http.StatusBadRequest)
		return
	}

	var req MessageReq
	if err = json.Unmarshal(body, &req); err != nil {
		log.Printf("请求解析内容失败，err=%v", err)
		common.ErrorStrResp(c, "请求失败！", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		log.Printf("请求内容为空，err=%v", err)
		common.ErrorStrResp(c, "请求失败！", http.StatusBadRequest)
		return
	}

	if msgs, ok := cache.GetFromCache(board); !ok {
		common.ErrorStrResp(c, "抱歉，由于服务器资源有限，当前剪切板数量已达上限，将为您自动跳转到public剪切板空间！", http.StatusBadRequest)
		return
	} else {
		isFile, fileName, fileType, base64Str := checkContentIsFile(req.Content)

		newMsg := &cache.Message{
			Content:  base64Str,
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			Ip:       realIp,
			Id:       fmt.Sprintf("%v", time.Now().UnixNano()), // 时间戳
			IsFile:   isFile,
			FileName: fileName,
			FileType: fileType,
		}

		msgs = append(msgs, newMsg)

		// msgs根据Time时间倒序排序
		sort.Slice(msgs, func(i, j int) bool {
			return msgs[i].Id > msgs[j].Id
		})

		// 只保留最新的10条数据
		if len(msgs) > MaxMessageSize {
			msgs = msgs[:MaxMessageSize]
		}

		cache.SetToCache(board, msgs, time.Hour*6)

		returnMsg := &cache.Message{
			Content:  "",
			Time:     newMsg.Time,
			Ip:       realIp,
			Id:       newMsg.Id, // 时间戳
			IsFile:   isFile,
			FileName: fileName,
			FileType: fileType,
		}
		if !isFile {
			returnMsg.Content = newMsg.Content
		}

		common.SuccessResp(c, BoardInfo{
			Board:    board,
			ExpireAt: cache.GetExpireAt(board),
			Messages: []*cache.Message{returnMsg},
		})
		return
	}
}
func GetMessage(c *gin.Context) {
	board := c.Param("board")
	if board == "" {
		common.ErrorStrResp(c, "board not found ！", http.StatusNotFound)
		return
	}

	id := c.Param("id")
	if id == "" {
		common.ErrorStrResp(c, "id not found ！", http.StatusNotFound)
		return
	}

	LogApiRequestIP(c, "GetMessage: "+board, -1)

	if msgs, ok := cache.GetFromCache(board); !ok {
		common.ErrorStrResp(c, "board not found ！", http.StatusNotFound)
		return
	} else {
		for _, msg := range msgs {
			if id == fmt.Sprintf("%v", msg.Id) {
				if msg.IsFile {
					// 解码 Base64 内容
					data, err := base64.StdEncoding.DecodeString(msg.Content)
					if err != nil {
						common.ErrorStrResp(c, "Failed to decode Base64 content!", http.StatusNotFound)
						return
					}
					c.Header("Content-Type", msg.FileType)
					c.Data(http.StatusOK, msg.FileType, data)
					return
				} else {
					c.Header("Content-Type", "text/plain; charset=utf-8")
					c.String(http.StatusOK, msg.FileType, msg.Content)
					return
				}
			}
		}
		common.ErrorStrResp(c, "message not found!", http.StatusNotFound)
		return
	}
}

func DeleteMessage(c *gin.Context) {
	board := c.Param("board")
	if board == "" {
		common.ErrorStrResp(c, "board not found ！", http.StatusNotFound)
		return
	}

	id := c.Param("id")
	if id == "" {
		common.ErrorStrResp(c, "id not found ！", http.StatusNotFound)
		return
	}

	LogApiRequestIP(c, "DeleteMessage: "+board, -1)

	if msgs, ok := cache.GetFromCache(board); !ok {
		common.ErrorStrResp(c, "board not found ！", http.StatusNotFound)
		return
	} else {
		removeIdx := -1
		for i, msg := range msgs {
			if id == fmt.Sprintf("%v", msg.Id) {
				removeIdx = i
			}
		}
		if removeIdx > -1 {
			msgs = append(msgs[:removeIdx], msgs[removeIdx+1:]...)
			if len(msgs) == 0 {
				cache.SetToCache(board, msgs, time.Minute*10)
			} else {
				cache.SetToCache(board, msgs, time.Hour*6)
			}
		}
		returnMsgs := make([]*cache.Message, 0)
		for _, msg := range msgs {
			if msg == nil {
				continue
			}
			returnMsg := &cache.Message{
				Content:  "",
				Time:     msg.Time,
				Ip:       msg.Ip,
				Id:       msg.Id, // 时间戳
				IsFile:   msg.IsFile,
				FileName: msg.FileName,
				FileType: msg.FileType,
			}
			if !msg.IsFile {
				returnMsg.Content = msg.Content
			}
			returnMsgs = append(returnMsgs, returnMsg)
		}
		common.SuccessResp(c, BoardInfo{
			Board:    board,
			ExpireAt: cache.GetExpireAt(board),
			Messages: returnMsgs,
		})
		return
	}

}

func checkContentIsFile(content string) (isFile bool, fileName, fileType, base64str string) {
	// 如果content符合正则表达式 ^(.*)#(data:.*;base64,.*)$
	re := regexp.MustCompile(`^(.*)#data:(.*);base64,(.*)$`)

	// 查找匹配的字符串
	matches := re.FindStringSubmatch(content)

	// 如果匹配，返回匹配的开头内容；否则，返回原始内容
	if len(matches) > 3 {
		return true, matches[1], matches[2], matches[3]
	}
	return false, "", "text/plain", content
}

func FetchBoard(c *gin.Context) {
	board := c.Param("board")
	if board == "" {
		common.ErrorStrResp(c, "Sorry, board not found!", http.StatusNotFound)
		return
	}

	realIp := LogApiRequestIP(c, "FetchBoard: "+board, -1)
	cache.SetBoardNameToCache(realIp, board, time.Hour*48)

	if msgs, ok := cache.GetFromCache(board); !ok {
		if board != "public" && cache.CacheSize() >= MaxBoardSize {
			common.ErrorStrResp(c, "抱歉，由于服务器资源有限，当前剪切板数量已达上限，将为您自动跳转到public剪切板空间！", http.StatusBadRequest)
			return
		}
		msgs = make([]*cache.Message, 0)
		cache.SetToCache(board, msgs, time.Minute*10)
		common.SuccessResp(c, BoardInfo{
			Board:    board,
			ExpireAt: cache.GetExpireAt(board),
			Messages: msgs,
		})
		return
	} else {
		returnMsgs := make([]*cache.Message, 0)
		for _, msg := range msgs {
			if msg == nil {
				continue
			}
			returnMsg := &cache.Message{
				Content:  "",
				Time:     msg.Time,
				Ip:       msg.Ip,
				Id:       msg.Id, // 时间戳
				IsFile:   msg.IsFile,
				FileName: msg.FileName,
				FileType: msg.FileType,
			}
			if !msg.IsFile {
				returnMsg.Content = msg.Content
			}
			returnMsgs = append(returnMsgs, returnMsg)
		}
		common.SuccessResp(c, BoardInfo{
			Board:    board,
			ExpireAt: cache.GetExpireAt(board),
			Messages: returnMsgs,
		})
		return
	}
}
