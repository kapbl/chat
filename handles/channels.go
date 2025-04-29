package handles

import (
	"errors"
	"kkj123/database"
	"kkj123/models"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
	curChannels := models.Channel{}
	if err := database.DB.Where("channel_id=?", channel).First(&curChannels).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.String(400, "该频道不存在")
		}
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
	curChannel := models.Channel{}
	database.DB.Where(models.Channel{ChannelID: rec.ChannelID, Name: rec.ChannelID}).FirstOrCreate(&curChannel)

	// 将用户与这个频道关联起来
	if claims, ok := ctx.Get("jwt_claims").(*JwtCustomClaims); ok {
		curUser := models.User{}
		if err := database.DB.Where("user_id=?", claims.UserID).First(&curUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ctx.JSON(401, map[string]string{"error": "从token中没有获取到userid"})
			}
		} else {
			// 关联用户和频道
			if err := database.DB.Model(&curUser).Association("Channels").Append(&curChannel); err != nil {
				log.Println("关联频道失败:", err)
			}
		}
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
	curChannel := models.Channel{}
	database.DB.Where(models.Channel{ChannelID: rec.ChannelID, Name: rec.ChannelID}).FirstOrCreate(&curChannel)
	// 将用户与这个频道关联起来
	if claims, ok := ctx.Get("jwt_claims").(*JwtCustomClaims); ok {
		curUser := models.User{}
		if err := database.DB.Where("user_id=?", claims.UserID).First(&curUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ctx.JSON(401, map[string]string{"error": "从token中没有获取到userid"})
			}
		} else {
			// 关联用户和频道
			if err := database.DB.Model(&curUser).Association("Channels").Append(&curChannel); err != nil {
				log.Println("关联频道失败:", err)
			}
		}
	}
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

// 当用户成功登陆后会自动列出已加入的频道
func InitJoinedChannels(ctx echo.Context) error {
	if claims, ok := ctx.Get("jwt_claims").(*JwtCustomClaims); ok {
		curUser := models.User{}
		if err := database.DB.Where("user_id=?", claims.UserID).First(&curUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ctx.JSON(401, map[string]string{"error": "从token中没有获取到userid"})
			}
		}
		// 用户id 查询 用户所有关联的频道
		curChannels := []models.Channel{}
		err := database.DB.Model(&curUser).Association("Channels").Find(&curChannels)
		if err != nil {
			log.Println("查询关联频道失败")
		}
		channelID := []string{}
		for _, v := range curChannels {
			channelID = append(channelID, v.ChannelID)
		}
		return ctx.JSON(200, map[string]interface{}{"channelIDs": channelID})
	}
	return nil
}

// 修改频道的名字
func ModifyChannel(ctx echo.Context) error {
	type ModifyChannelRequest struct {
		ChannelID string `json:"channelId" binding:"required,min=2,max=50"`
		NewName   string `json:"newName" binding:"required,min=2,max=50"`
	}
	var req ModifyChannelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(400, map[string]string{"error": "数据格式不对"})
	}
	var existingChannel models.Channel
	if err := database.DB.Where("channel_id=?", req.ChannelID).First(&existingChannel).Error; err != nil {
		ctx.JSON(400, map[string]string{"error": "数据库中没有这个频道"})
		return err
	}
	var nameCheck models.Channel
	if database.DB.Where("channel_id=?", req.NewName).First(&nameCheck).RowsAffected > 0 {
		ctx.JSON(400, map[string]string{"message": "频道名称已存在"})
		return errors.New("频道名称已存在")
	}
	updateData := models.Channel{
		ChannelID: req.NewName,
		Name:      req.NewName,
	}
	if err := database.DB.Model(&models.Channel{}).Where("channel_id = ?", req.ChannelID).Updates(updateData).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "频道名称修改失败"})
		return err
	}
	return ctx.JSON(200, map[string]string{"message": "频道修改成功"})
}

type DeleteChannelRequest struct {
	ChannelID string `json:"channelID" binding:"required"`
}

// 用户删除一个已经加入的频道
func DeleteChannel(ctx echo.Context) error {
	var req DeleteChannelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求格式不对"})
		return err
	}
	// 判断这个频道是否存在
	var channel models.Channel
	if err := database.DB.Where("channel_id=?", req.ChannelID).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(404, map[string]string{"message": "频道不存在"})
			return err
		}
	}
	// 判断这个用户是否关联了这个频道
	curUser := models.User{}
	if claims, ok := ctx.Get("jwt_claims").(*JwtCustomClaims); ok {
		if err := database.DB.Where("user_id=?", claims.UserID).First(&curUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(400, map[string]string{"error": "未找到该用户"})
			}
		}
		// 找到了该用户
		// 联合查找
		var curChannels []models.Channel
		if err := database.DB.Model(&curUser).Association("Channels").Find(&curChannels); err != nil {
			ctx.JSON(400, map[string]string{"error": "该用户未与频道关联"})
		}

		for _, c := range curChannels {
			if c.ChannelID == channel.ChannelID {
				// 取消关联
				database.DB.Model(&curUser).Association("Channels").Delete(&c)
			}
		}
	}
	// 如果频道存在，执行删除操作
	if err := database.DB.Delete(&channel).Error; err != nil {
		ctx.JSON(500, map[string]string{"message": "删除失败"})
		return err
	}
	return nil
}
