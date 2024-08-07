package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ua-parser/uap-go/uaparser"
	"hash/fnv"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	// 等级列表
	levels = []string{"青铜Ⅰ", "青铜Ⅱ", "青铜Ⅲ", "白银Ⅰ", "白银Ⅱ", "白银Ⅲ", "黄金Ⅰ", "黄金Ⅱ", "黄金Ⅲ", "黄金Ⅳ",
		"铂金Ⅰ", "铂金Ⅱ", "铂金Ⅲ", "铂金Ⅳ", "钻石Ⅰ", "钻石Ⅱ", "钻石Ⅲ", "钻石Ⅳ", "钻石Ⅴ",
		"星耀Ⅰ", "星耀Ⅱ", "星耀Ⅲ", "星耀Ⅳ", "星耀Ⅴ", "最强王者", "无双王者", "荣耀王者", "传奇王者"}
	// 英雄列表
	heroes = []string{"廉颇", "小乔", "赵云", "墨子", "妲己", "嬴政", "孙尚香", "鲁班七号", "庄周", "刘禅", "高渐离",
		"阿轲", "钟无艳", "孙膑", "扁鹊", "白起", "芈月", "吕布", "周瑜", "夏侯惇", "甄姬", "曹操", "典韦", "宫本武藏",
		"李白", "马可波罗", "狄仁杰", "达摩", "项羽", "武则天", "老夫子", "关羽", "貂蝉", "安琪拉", "程咬金", "露娜", "姜子牙",
		"刘邦", "韩信", "王昭君", "兰陵王", "花木兰", "张良", "不知火舞", "娜可露露", "橘右京", "亚瑟", "孙悟空", "牛魔", "后羿",
		"刘备", "张飞", "李元芳", "虞姬", "钟馗", "成吉思汗", "杨戬", "雅典娜", "蔡文姬", "太乙真人", "哪吒", "诸葛亮", "黄忠",
		"大乔", "东皇太一", "干将莫邪", "鬼谷子", "铠", "百里守约", "百里玄策", "苏烈", "梦奇", "女娲", "明世隐", "公孙离",
		"杨玉环", "裴擒虎", "弈星", "狂铁", "米莱狄", "元歌", "孙策", "司马懿", "盾山", "伽罗", "沈梦溪", "李信", "上官婉儿",
		"嫦娥", "猪八戒", "盘古", "瑶", "云中君", "曜", "马超", "西施", "鲁班大师", "蒙犽", "镜", "蒙恬", "阿古朵", "夏洛特",
		"澜", "司空震", "艾琳", "云缨", "金蝉", "暃", "桑启", "戈娅", "海月", "赵怀真", "莱西奥", "姬小满", "亚连", "朵莉亚",
		"海诺", "敖隐", "大司命"}
)

type PeerName struct {
	model       string `json:"model"`
	os          string `json:"os"`
	browser     string `json:"browser"`
	deviceType  string `json:"type"`
	deviceName  string `json:"deviceName"`
	displayName string `json:"displayName"`
}

type Peer struct {
	socket       *websocket.Conn `json:"-"`
	ip           string          `json:"-"`
	id           string          `json:"id"`
	rtcSupported bool            `json:"rtcSupported"`
	name         *PeerName       `json:"name"`
	board        string          `json:"board"`
	lastBeat     time.Time       `json:"-"`
	timer        *time.Timer     `json:"-"`

	cancelKeepAlive chan struct{} `json:"-"`
	mu              sync.Mutex    `json:"-"`
}

type PeerServer struct {
	upgrader websocket.Upgrader
	rooms    map[string]map[string]*Peer
	boards   map[string]map[string]map[string]bool
	mu       sync.Mutex
}

// NewPeer creates a new Peer
func NewPeer(socket *websocket.Conn, c *gin.Context) *Peer {
	newPeer := &Peer{
		socket:          socket,
		cancelKeepAlive: make(chan struct{}),
	}

	// set ip
	newPeer.ip = LogApiRequestIP(c, "NewPeer", -1)
	// if ip is localhost, set it to 127.0.0.1
	if newPeer.ip == "::1" || newPeer.ip == "::ffff:127.0.0.1" {
		newPeer.ip = "127.0.0.1"
	}
	// peerId由PeerServer生成，写入Cookie
	if peerId, err := c.Cookie("peerid"); err == nil {
		newPeer.id = peerId
	}
	// set rtcSupported
	if strings.Index(c.Request.URL.String(), "webrtc") > -1 {
		newPeer.rtcSupported = true
	} else {
		newPeer.rtcSupported = false
	}
	// set name
	uaString := c.GetHeader("User-Agent")
	parser := uaparser.NewFromSaved()
	client := parser.Parse(uaString)

	deviceName := ""

	if client.Os.Family != "" {
		deviceName = client.Os.Family
		if client.Os.Family == "Mac OS X" {
			deviceName = "Mac"
		}
		deviceName += " "
	}

	if client.UserAgent.Family != "" {
		if strings.Contains(client.UserAgent.Family, "WKWebView") {
			deviceName += "WKWebView"
		} else if strings.HasPrefix(client.UserAgent.Family, "Chrome Mobile") {
			deviceName += "Mobile Chrome"
		} else {
			deviceName += client.UserAgent.Family
		}
	} else {
		deviceName += client.Device.Model
	}

	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	displayName := getRandomHero(newPeer.id)

	newPeer.name = &PeerName{
		model:       client.Device.Model,     // 设备型号
		os:          client.Os.Family,        // 操作系统
		browser:     client.UserAgent.Family, // 浏览器
		deviceType:  client.Device.Family,    // 设备类型
		deviceName:  deviceName,              // 显示设备名称
		displayName: displayName,             // 显示名称
	}

	newPeer.lastBeat = time.Now()

	return newPeer
}

func (p *Peer) getInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           p.id,
		"ip":           p.ip,
		"rtcSupported": p.rtcSupported,
		"name": map[string]interface{}{
			"model":       p.name.model,
			"os":          p.name.os,
			"browser":     p.name.browser,
			"deviceType":  p.name.deviceType,
			"deviceName":  p.name.deviceName,
			"displayName": p.name.displayName,
		},
	}
}

// NewPeerServer creates a new PeerServer
func NewPeerServer() *PeerServer {
	return &PeerServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		rooms:  make(map[string]map[string]*Peer),           // room -> id -> peer
		boards: make(map[string]map[string]map[string]bool), // board -> room -> id -> bool
	}
}

// HandleConnection handles a new peer connection
func (s *PeerServer) HandleConnection(c *gin.Context) {
	// Check if peerid cookie exists, if not generate a new peerId
	var peerId string
	// Check if peerid cookie exists
	if cookie, err := c.Request.Cookie("peerid"); err == nil && cookie.Value != "" {
		//log.Println("Cookie peerid found:", cookie.Value)
		peerId = cookie.Value
	} else {
		// Generate a new peerId if cookie doesn't exist
		peerId = uuid.NewString()
		c.Header("Set-Cookie", "peerid="+peerId+";SameSite=Strict;Secure")
		//log.Println("Set Cookie peerid:", peerId)
	}

	// Upgrade the connection to a websocket connection
	socket, err := s.upgrader.Upgrade(c.Writer, c.Request, c.Writer.Header())
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer socket.Close()

	// Create a new Peer instance
	peer := NewPeer(socket, c)
	peer.id = peerId
	s.joinRoom(peer)

	// Start a goroutine to keep the connection alive
	go s.keepAlive(peer)

	// Send a display-name message to the peer
	s.send(peer, map[string]interface{}{
		"type": "display-name",
		"message": map[string]string{
			"displayName": peer.name.displayName,
			"deviceName":  peer.name.deviceName,
		},
	})

	// Read messages from the socket
	for {
		_, message, err := socket.ReadMessage()
		if err != nil {
			//log.Println("Read error:", err)
			s.leaveRoom(peer)
			break
		}
		// Handle the received message
		s.handleMessage(peer, message)
	}

	// Cancel the keep alive goroutine
	s.cancelKeepAlive(peer)
}

func (s *PeerServer) joinRoom(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if room doesn't exist, create it
	if _, exists := s.rooms[peer.ip]; !exists {
		s.rooms[peer.ip] = make(map[string]*Peer)
	}

	// add peer to room
	s.rooms[peer.ip][peer.id] = peer
	log.Printf("Peer joined: %s (ID: %s, Board: %s)", peer.ip, peer.id, peer.board)

	// Notify other peers in the room
	for _, otherPeer := range s.rooms[peer.ip] {
		if otherPeer.id != peer.id {
			s.send(otherPeer, map[string]interface{}{
				"type": "peer-joined",
				"peer": peer.getInfo(),
			})
		}
	}

	// Send current peers to the new peer
	peers := make([]map[string]interface{}, 0, len(s.rooms[peer.ip])-1)
	for _, otherPeer := range s.rooms[peer.ip] {
		if otherPeer.id != peer.id {
			peers = append(peers, otherPeer.getInfo())
		}
	}
	s.send(peer, map[string]interface{}{
		"type":  "peers",
		"peers": peers,
	})
}

func (s *PeerServer) leaveRoom(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// remove peer from room
	if room, exists := s.rooms[peer.ip]; exists {
		if _, peerExists := room[peer.id]; !peerExists {
			return
		}
		s.cancelKeepAlive(peer)
		peer.socket.Close()
		delete(room, peer.id)
		log.Printf("Peer left: %s (ID: %s, Board: %s)", peer.ip, peer.id, peer.board)

		if len(room) == 0 {
			// if room is empty, remove it
			delete(s.rooms, peer.ip)
		} else {
			// notify all other peers
			for _, otherPeer := range room {
				s.send(otherPeer, map[string]interface{}{
					"type":   "peer-left",
					"peerId": peer.id,
				})
			}
		}
	}

	if peer.board != "" {
		delete(s.boards[peer.board][peer.ip], peer.id)
		if len(s.boards[peer.board][peer.ip]) == 0 {
			delete(s.boards[peer.board], peer.ip)
		}
		if len(s.boards[peer.board]) == 0 {
			delete(s.boards, peer.board)
		}
	}
}

func (s *PeerServer) handleMessage(sender *Peer, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	switch msg["type"] {
	case "disconnect":
		s.leaveRoom(sender)
	case "pong":
		sender.lastBeat = time.Now()
		board, _ := msg["board"].(string)
		sender.board = board
		s.mu.Lock()
		if _, exists := s.boards[board]; !exists {
			s.boards[board] = make(map[string]map[string]bool)
		}
		if _, exists := s.boards[board][sender.ip]; !exists {
			s.boards[board][sender.ip] = make(map[string]bool)
		}
		s.boards[board][sender.ip][sender.id] = true
		log.Printf("Receive pong from board=%s, ip=%v, id=%v", board, sender.ip, sender.id)
		s.mu.Unlock()
	case "board-update":
		sender.lastBeat = time.Now()
		board, _ := msg["board"].(string)
		//log.Printf("Receive board-update from board=%s, ip=%v, id=%v", board, sender.ip, sender.id)
		if peers, exists := s.boards[board]; exists {
			for ip, ids := range peers {
				if room, exists := s.rooms[ip]; exists {
					for id := range ids {
						if id != sender.id {
							s.send(room[id], map[string]interface{}{
								"type":  "board-update",
								"board": board,
							})
						}
					}
				}
			}
		}
	}

	// RTC message tp specified peer
	if to, exists := msg["to"]; exists {
		recipientId, _ := to.(string)
		if room, exists := s.rooms[sender.ip]; exists {
			if recipient, exists := room[recipientId]; exists {
				msg["sender"] = sender.id
				delete(msg, "to")
				s.send(recipient, msg)
			}
		}
	}
}

func (s *PeerServer) send(peer *Peer, message map[string]interface{}) {
	if peer == nil {
		return
	}

	peer.mu.Lock()
	defer peer.mu.Unlock()

	if err := peer.socket.WriteJSON(message); err != nil {
		log.Printf("Write msssage:%v error:%v", message, err)
	}
}

func (s *PeerServer) keepAlive(peer *Peer) {
	s.send(peer, map[string]interface{}{"type": "ping", "board": peer.board})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if peer.lastBeat.Add(60 * time.Second).Before(time.Now()) {
				s.leaveRoom(peer)
				return
			}
			s.send(peer, map[string]interface{}{"type": "ping", "board": peer.board})
		case <-peer.cancelKeepAlive:
			//log.Println("KeepAlive canceled for peer:", peer.id)
			return
		}
	}
}

func (s *PeerServer) cancelKeepAlive(peer *Peer) {
	if peer != nil && peer.cancelKeepAlive != nil && !isClosed(peer.cancelKeepAlive) {
		close(peer.cancelKeepAlive)
	}
}

// seededRandom is a function that implements the Linear Congruential Generator (LCG) algorithm.
// It takes a seed value as input and returns a float64 pseudo-random number.
func seededRandom(seed uint32) float64 {
	const a = 1664525
	const c = 1013904223

	seed = a*seed + c
	return float64(seed) / float64(1<<32)
}

// hashStringToSeed hashes a string to a seed value.
func hashStringToSeed(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// getRandomHero generates a random hero and level.
// It takes a seed string as input and returns a string representing a hero and level.
func getRandomHero(seedStr string) string {
	heroSeed := hashStringToSeed(seedStr)
	levelSeed := heroSeed + 1 // 加1以确保种子不同
	heroIndex := int(math.Floor(seededRandom(heroSeed) * float64(len(heroes))))
	levelIndex := int(math.Floor(seededRandom(levelSeed) * float64(len(levels))))
	return fmt.Sprintf("%s %s", levels[levelIndex], heroes[heroIndex])
}

// isClosed checks if the given channel is closed.
// It returns true if the channel is closed, otherwise false.
func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
