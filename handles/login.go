package handles

import (
	"kkj123/database"
	"kkj123/models"

	"github.com/labstack/echo/v4"
)

func Login(ctx echo.Context) error {
	currentUser := models.User{}
	if err := ctx.Bind(&currentUser); err != nil {
		return ctx.JSON(400, map[string]string{"error": "用户或者密码错误"})
	}
	// 查询数据库
	if err := database.DB.Where("username = ? AND password = ?", currentUser.Username, currentUser.Password).First(&currentUser).Error; err != nil {
		return ctx.JSON(400, map[string]string{"error": "用户或者密码错误"})
	}

	ctx.JSON(200, map[string]string{"message": "success"})
	return nil
}
