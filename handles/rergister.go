package handles

import (
	"kkj123/database"
	"kkj123/models"

	"github.com/labstack/echo/v4"
)

func Register(ctx echo.Context) error {
	currUser := models.User{}
	if err := ctx.Bind(&currUser); err != nil {
		ctx.JSON(400, map[string]string{"message": "请求参数错误"})
		return nil
	}
	if currUser.Username == "" || currUser.Password == "" {
		ctx.JSON(400, map[string]string{"message": "用户名或密码不能为空"})
		return nil
	}
	// 将用户信息存入数据库
	if err := database.DB.Where("username = ?", currUser.Username).First(&currUser).Error; err == nil {
		return ctx.JSON(400, map[string]string{"message": "用户名已存在"})
	}
	// 创建用户
	database.DB.Create(&currUser)
	ctx.JSON(200, map[string]string{"message": "注册成功"})
	return nil
}
