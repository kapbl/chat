package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"kkj123/database"
	"kkj123/models"
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
	channels  = map[string][]*client{
		"bot": make([]*client, 0),
	}
)

func InitChannels() {
	allChannels := []models.Channel{}
	database.DB.Find(&allChannels)
	for _, v := range allChannels {
		channels[v.Name] = make([]*client, 0)
	}
}

// 在 main 或初始化部分启动 Redis 频道的订阅
func IninDafaultChannel() {
	go subscribeRedisChannel("bot")
}

func subscribeRedisChannel(channel string) {
	ctx := context.Background()
	pubsub := database.RedisDB.Subscribe(ctx, channel)
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		// 当接收到消息时，广播给所有连接的 WebSocket 客户端
		clientsMu.Lock()
		for _, c := range channels[channel] {
			if msg.Payload[11:26] == c.user_id {
				continue
			}
			log.Println("广播消息", msg.Payload[11:26])
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				c.conn.Close()
			}
		}
		clientsMu.Unlock()
	}
}

// websocket 路由处理函数
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
	for {
		chatMsg := ChatMessage{}
		err := ws.ReadJSON(&chatMsg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// 将消息发布到 Redis
		if err := publishMessage("bot", chatMsg); err != nil {
			log.Println("发布消息失败:", err)
		}
	}
	return nil
}

func registerClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// !step1: 获取连接的初始信息
	currUser := struct {
		User string `json:"user"`
	}{}
	ws.ReadJSON(&currUser)

	claims := JWTUnencoder([]byte("my_secret"), currUser.User)
	if claims == nil {
		return
	}
	// !step2: 在服务器环境中建立该用户
	userInstance := client{
		conn:    ws,
		user_id: claims.UserID,
		status:  online,
	}

	// * map 像这样： 14ae3dd733491c76bd64b942a99acd49b9f2d7ac7f1e74b516f7233443f05af6 = client{}
	clients[userInstance.user_id] = &userInstance
	// 从数据库中查询该用户
	targetUser := models.User{}
	targetChannel := []models.Channel{}
	database.DB.Where("user_id=?", userInstance.user_id).First(&targetUser)
	if err := database.DB.Model(&targetUser).Association("Channels").Find(&targetChannel); err != nil {
		log.Println(err)
	}
	// 将该用户加入到自己已经保存的频道
	for _, v := range targetChannel {
		channels[v.ChannelID] = append(channels[v.ChannelID], &userInstance)
	}
	log.Println("客户端连接 ", targetUser.Username)
	// 发送欢迎消息
	welcomeMsg := ChatMessage{
		Sender:  "bot",
		Content: ws.RemoteAddr().String(),
		Channel: "bot",
	}
	if err := ws.WriteJSON(welcomeMsg); err != nil {
		log.Println("欢迎消息发送失败:", err)
	}
}

func publishMessage(channel string, message ChatMessage) error {
	// 序列化 ChatMessage 为 JSON
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}
	// 发布消息到 Redis
	ctx := context.Background()
	err = database.RedisDB.Publish(ctx, channel, messageData).Err()
	if err != nil {
		return fmt.Errorf("消息发布失败: %v", err)
	}
	return nil
}

func unregisterClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// 删除全局客户端
	var targetClient *client
	for addr, c := range clients {
		if c.conn == ws {
			delete(clients, addr)
			targetClient = c
			break
		}
	}
	if targetClient == nil {
		return
	}
	// 从所有频道中移除
	for channel := range channels {
		for i, c := range channels[channel] {
			if c.user_id == targetClient.user_id {
				channels[channel] = append(channels[channel][:i], channels[channel][i+1:]...)
				break
			}
		}
	}
}
