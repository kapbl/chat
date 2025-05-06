package handles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kkj123/database"
	"kkj123/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type client struct {
	conn    *websocket.Conn // websocket 连接对象
	user_id string          // 数据库中唯一的用户id
}

type clientList struct {
	sync.RWMutex
	clients []*client
}

var channels sync.Map

// 添加一个用户
func AddClient(key string, c *client) {
	v, ok := channels.Load(key)
	if ok {
		clist := v.(*clientList)
		clist.Lock()
		clist.clients = append(clist.clients, c)
		clist.Unlock()
		return
	}
	// 首次创建表
	newList := &clientList{
		clients: []*client{c},
	}

	actual, loader := channels.LoadOrStore(key, newList)
	if loader {
		clist := actual.(*clientList)
		clist.Lock()
		clist.clients = append(clist.clients, c)
		clist.Unlock()
	}
}

// 获取客户端列表，安全读
func GetClients(key string) []*client {
	v, ok := channels.Load(key)
	if !ok {
		return nil
	}
	list := v.(*clientList)
	list.RLock()
	defer list.RUnlock()

	ret := make([]*client, len(list.clients))
	copy(ret, list.clients)
	return ret
}

// 删除制定的客户端
func DeleteClient(key string, targetID string) {
	v, ok := channels.Load(key)
	if !ok {
		return
	}

	list := v.(*clientList)
	list.Lock()
	defer list.Unlock()

	// 遍历查找目标
	for i, c := range list.clients {
		if c.user_id == targetID {
			// 快速删除（不保持顺序）
			list.clients[i] = list.clients[len(list.clients)-1]
			list.clients = list.clients[:len(list.clients)-1]
			break
		}
	}

	// 自动清理空列表
	if len(list.clients) == 0 {
		channels.Delete(key)
	}
}

type ChatMessage struct {
	User    string `json:"user"`
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

// 服务器初始环境中订阅所有的已经存在的频道
func InitChannelsSubscribeRedisChannels() {
	// 订阅频道
	allChannels := []models.Channel{}
	database.DB.Find(&allChannels)
	for _, v := range allChannels {
		go subscribeRedisChannel(v.ChannelID)
	}
}

// 订阅频道
func subscribeRedisChannel(channel string) {
	pubsub := database.RedisDB.Subscribe(context.Background(), channel)
	defer pubsub.Close()
	type PartialMessage struct {
		User string `json:"user"`
	}
	for msg := range pubsub.Channel() {
		var sender PartialMessage
		err := json.Unmarshal([]byte(msg.Payload), &sender)
		if err != nil {
			fmt.Println("JSON 解析失败:", err)
			return
		}
		claims := JWTUnencoder([]byte("my_secret"), sender.User)
		if claims == nil {
			log.Println("claims不能为空")
		}
		targetChannel := GetClients(msg.Channel)
		for _, c := range targetChannel {
			if c.user_id == claims.UserID {
				continue
			}
			log.Println("广播消息", claims.UserID)

			err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Printf("Error sending message to client: %v", err)
				c.conn.Close()
			}
		}
	}
}

// websocket 路由处理函数
func HandleWebSocker(ctx echo.Context) error {
	// 升级到websocket协议
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Println("WebSocket 升级错误:", err)
		return err
	}
	defer ws.Close()

	// 注册客户端
	registerClient(ws)
	defer unregisterClient(ws)
	// 处理消息循环
	messageLoop(ws)
	return nil
}

func messageLoop(ws *websocket.Conn) {
	for {
		chatMsg := ChatMessage{}
		err := ws.ReadJSON(&chatMsg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// 将消息发布到 Redis
		if err := publishMessage(chatMsg.Channel, chatMsg); err != nil {
			log.Println("发布消息失败:", err)
		}
	}
}

// 在服务器环境中注册一个用户实例
func registerClient(ws *websocket.Conn) error {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// !step1: 获取连接的初始信息
	currUser := struct {
		User string `json:"user"`
	}{}
	if err := ws.ReadJSON(&currUser); err != nil {
		return errors.New("json解析失败")
	}

	// !step2: 解析实际的user_id
	claims := JWTUnencoder([]byte("my_secret"), currUser.User)
	if claims == nil {
		return errors.New("claim不能为空")
	}

	// !step3: 在服务器环境中建立该用户
	userInstance := client{
		conn:    ws,
		user_id: claims.UserID,
	}

	clients[userInstance.user_id] = &userInstance

	// !step4: 从数据库中查询该用户
	targetUser := models.User{}
	targetChannel := []models.Channel{}
	if err := database.DB.Where("user_id=?", userInstance.user_id).First(&targetUser).Error; err != nil {
		return err
	}
	if err := database.DB.Model(&targetUser).Association("Channels").Find(&targetChannel); err != nil {
		return err
	}

	// !step5: 将该用户加入到自己已经保存的频道
	for _, v := range targetChannel {
		AddClient(v.ChannelID, &userInstance)
	}

	// !step6: 发送欢迎消息
	log.Println("客户端已经连接 ", targetUser.Username)
	welcomeMsg := ChatMessage{
		User:    "bot",
		Content: targetUser.UserID,
		Channel: "bot",
	}
	if err := ws.WriteJSON(welcomeMsg); err != nil {
		log.Println("欢迎消息发送失败:", err)
	}
	return nil
}

// 发布一个消息
func publishMessage(channel string, message ChatMessage) error {
	// 序列化 ChatMessage 为 JSON
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}
	// 发布消息到 Redis
	err = database.RedisDB.Publish(context.Background(), channel, messageData).Err()
	if err != nil {
		return fmt.Errorf("消息发布失败: %v", err)
	}
	return nil
}

// 在 服务器环境中移除一个用户实例
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
}
