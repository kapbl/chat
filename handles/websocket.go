package handles

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Status string

const (
	online  Status = "online"
	offline Status = "offline"
)

type client struct {
	conn    *websocket.Conn
	user_id string
	status  Status
}

type ChatMessage struct {
	Sender  string `json:"sender"`
	Channel string `json:"channel"` // 频道
	Content string `json:"content"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[string]*client)
	clientsMu sync.Mutex
)
var channels = make(map[string][]*client) // 频道列表

func init() {
	// 初始化频道列表， 测试用的
	channels["bot"] = make([]*client, 0)
	channels["game"] = make([]*client, 0)
	channels["sport"] = make([]*client, 0)
	channels["book"] = make([]*client, 0)
}

func HandleWebSocker(ctx echo.Context) error {
	// 升级到websocket协议
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
	}
	defer ws.Close()
	// 注册客户端
	registerClient(ws)
	defer unregisterClient(ws)
	// 处理消息循环
	messageHandler(ws)
	return nil
}

// 注册一个客户端
func registerClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[ws.RemoteAddr().String()] = &client{
		conn:    ws,
		user_id: ws.RemoteAddr().String(),
		status:  online,
	}
	// 将客户端添加到频道列表，测试用
	channels["bot"] = append(channels["bot"], clients[ws.RemoteAddr().String()])
	log.Println("Client connected:", ws.RemoteAddr())
	for k := range clients {
		log.Println("current clients:", k)
	}
	go func() {
		msg := ChatMessage{
			Sender:  "bot",
			Content: ws.RemoteAddr().String(),
			Channel: "bot",
		}
		err := ws.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			ws.Close()
		}
	}()
}

// 注销一个客户端
func unregisterClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, c := range clients {
		if c.conn == ws {
			delete(clients, ws.RemoteAddr().String())
			break
		}
	}
	log.Println("Client disconnected:", ws.RemoteAddr())
}

func messageHandler(ws *websocket.Conn) {
	for {
		chatMsg := ChatMessage{}
		err := ws.ReadJSON(&chatMsg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// 判断是否是频道消息
		if chatMsg.Channel != "" {
			handleChannelMessage(chatMsg)
		}
	}
}

// 处理频道消息
func handleChannelMessage(message ChatMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if _, exist := channels[message.Channel]; !exist {
		log.Println("Channel not found:", message.Channel)
		return
	}
	for _, c := range channels[message.Channel] {
		if clients[c.user_id] == nil {
			log.Println("Client not found:", c.user_id)
			continue
		}
		if clients[c.user_id].status == offline {
			log.Println("Client is offline:", c.user_id)
			continue
		}
		if clients[c.user_id].user_id == message.Sender {
			continue
		}
		err := clients[c.user_id].conn.WriteJSON(message)
		if err != nil {
			log.Println("Write error:", err)
			clients[c.user_id].conn.Close()
		}
	}
	log.Println("Private message sent to:", message.Channel, "Content:", message.Content)
}
