package handles

import (
	"kkj123/database"
	"kkj123/models"
	"log"

	"github.com/labstack/echo/v4"
)

// 频道搜索路由处理函数
func ChannelSearch(ctx echo.Context) error {
	// 获取频道名称
	channel := ctx.QueryParam("channel")
	if channel == "" {
		return ctx.String(400, "channel is required")
	}
	// 获取频道列表
	clientsMu.Lock()
	defer clientsMu.Unlock()
	// 根据频道名称查询，原有的频道的名称是否包含

	if _, ok := channels[channel]; !ok {
		return ctx.String(404, "channel not found")
	}
	type Channel struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Members int    `json:"members"`
		Unread  int    `json:"unread"`
	}
	results := []Channel{
		{ID: channel, Name: channel, Members: 42},
	}
	// 返回频道列表
	return ctx.JSON(200, results)
}

// 加入频道路由处理函数
func JoinChannel(ctx echo.Context) error {
	// 获取频道名称
	type mess struct {
		Sender    string `json:"sender"`    // 发送者
		ChannelID string `json:"channelID"` // 频道id
	}
	rec := mess{}
	if err := ctx.Bind(&rec); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
	}
	// 加入频道
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if _, ok := channels[rec.ChannelID]; !ok {
		ctx.JSON(400, map[string]string{"message": "频道不存在"})
	}
	channels[rec.ChannelID] = append(channels[rec.ChannelID], clients[rec.Sender])
	curChannel := []models.Channel{
		{ChannelID: rec.ChannelID, Name: rec.ChannelID},
	}
	if err := database.DB.Create(&curChannel).Error; err != nil {
		log.Fatal("创建频道失败:", err)
	}
	// 发送消息
	msg := ChatMessage{
		Sender:  "bot",
		Content: "欢迎加入频道 " + rec.ChannelID,
		Channel: rec.ChannelID,
	}
	err := clients[rec.Sender].conn.WriteJSON(msg)
	if err != nil {
		log.Println("Write error:", err)
		clients[rec.Sender].conn.Close()
	}
	return ctx.String(200, "join channel success")
}

// 创建频道路由处理函数
func CreateChannel(ctx echo.Context) error {
	type mess struct {
		Sender             string `json:"sender"`             // 发送者
		ChannelID          string `json:"channelID"`          // 频道id
		ChannelDescription string `json:"channelDescription"` // 频道描述
	}
	rec := mess{}
	if err := ctx.Bind(&rec); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
	}
	// 创建频道
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if _, ok := channels[rec.ChannelID]; ok {
		ctx.JSON(400, map[string]string{"message": "频道已存在"})
	}
	channels[rec.ChannelID] = make([]*client, 0)
	// 将当前用户加入频道
	channels[rec.ChannelID] = append(channels[rec.ChannelID], clients[rec.Sender])
	// 将该频道存入用户数据库中的channels字段中

	// 发送消息
	clients[rec.Sender].conn.WriteJSON(ChatMessage{
		Sender:  "bot",
		Content: "欢迎加入频道 " + rec.ChannelID,
		Channel: rec.ChannelID,
	})
	// 订阅该频道
	go subscribeRedisChannel(rec.ChannelID)
	return ctx.String(200, "create channel success")
}
