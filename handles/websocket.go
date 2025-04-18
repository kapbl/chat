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
	Username string `json:"username"`
	Content  string `json:"content"`
	TargetID string `json:"target_id"`
	GroupID  string `json:"group_id"` // 群组id
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[string]*client)
	clientsMu sync.Mutex
)

func HandleWebSocker(ctx echo.Context) error {
	// 升级到websocket协议
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return err
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
	log.Println("Client connected:", ws.RemoteAddr())
	go func() {
		msg := ChatMessage{
			Username: "system",
			Content:  ws.RemoteAddr().String(),
			TargetID: ws.RemoteAddr().String(),
			GroupID:  "",
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
		// 判断是否是私人消息
		if chatMsg.TargetID != "" {
			handlePrivateMessage(chatMsg)
		} else if chatMsg.GroupID != "" {
			broadcastMessageGroup(chatMsg)
		}
	}
}

// 处理群组消息
func broadcastMessageGroup(message ChatMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	log.Println(groups[message.GroupID])
	if _, ok := groups[message.GroupID]; !ok {
		log.Println("Group not found:", message.GroupID)
		return
	}
	for _, c := range groups[message.GroupID].Inclient {
		if clients[c] == nil {
			log.Println("Client not found:", c)
			continue
		}
		if clients[c].status == offline {
			log.Println("Client is offline:", c)
			continue
		}
		err := clients[c].conn.WriteJSON(message)
		if err != nil {
			log.Println("Write error:", err)
			clients[c].conn.Close()
		}
	}
}

// 广播消息给所有客户端
func broadcastMessage(message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, c := range clients {
		if c.status == offline {
			continue
		}
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Write error:", err)
			c.conn.Close()
		}
	}
}

// 处理私人消息
// 发送给指定的客户端
func handlePrivateMessage(message ChatMessage) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if clients[message.TargetID] == nil {
		log.Println("Target client not found:", message.TargetID)
	} else if clients[message.TargetID].status == offline {
		log.Println("Target client is offline:", message.TargetID)
		return
	}
	clients[message.TargetID].conn.WriteJSON(message)
	log.Println("Private message sent to:", message.TargetID, "Content:", message.Content)
}
